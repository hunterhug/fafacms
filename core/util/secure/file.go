package secure

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// FileChunkSize 文件分块明文大小（1 MiB）。
//
// 分块加密的意义：
//  1. 可流式解密，不必把整个文件读进内存；
//  2. 大文件不受 GCM 单条消息长度限制；
//  3. 密文偏移可精确计算，从而支持 HTTP Range（音视频拖动）。
const FileChunkSize = 1 << 20

// FileNonceSize 文件加密的随机前缀长度（字节）。
// 实际 nonce = baseNonce(8) ‖ counter(4)，共 12 字节。
const FileNonceSize = 8

// NewFileNonce 生成文件加密用的随机前缀。
func NewFileNonce() ([]byte, error) {
	b := make([]byte, FileNonceSize)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// DeriveFileNonce 由 base nonce 派生一条**用途隔离**的独立 nonce。
//
// 为什么必须派生而不能复用：GCM 的密文是明文与密钥流的异或，而密钥流只由 (key, nonce)
// 决定（AAD 只影响 tag）。若用同一 (FEK, baseNonce) 加密两份不同内容（例如原图与缩略图），
// 会直接泄漏两份明文的异或值，并破坏认证。故衍生内容必须使用不同的 nonce。
//
// 派生是确定性的，因此下载侧可用同样的 purpose 重新算出来。
func DeriveFileNonce(baseNonce []byte, purpose string) ([]byte, error) {
	if len(baseNonce) != FileNonceSize {
		return nil, fmt.Errorf("secure: nonce 长度应为 %d", FileNonceSize)
	}
	return HKDF(baseNonce, nil, []byte("hunterhug:file:"+purpose), FileNonceSize), nil
}

// ThumbNonce 返回缩略图所用的 nonce（由原图 nonce 派生）。
func ThumbNonce(baseNonce []byte) ([]byte, error) {
	return DeriveFileNonce(baseNonce, "thumb")
}

// FileNonceHex 把 base nonce 编码为可入库的十六进制字符串。
func FileNonceHex(baseNonce []byte) string {
	return hex.EncodeToString(baseNonce)
}

// ParseFileNonce 解析库中存的 nonce。
func ParseFileNonce(s string) ([]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("secure: nonce 编码无效: %w", err)
	}
	if len(b) != FileNonceSize {
		return nil, fmt.Errorf("secure: nonce 长度应为 %d，实际 %d", FileNonceSize, len(b))
	}
	return b, nil
}

// chunkNonce 由 base nonce 与块序号构造该块的 nonce。
func chunkNonce(baseNonce []byte, index int) []byte {
	n := make([]byte, 12)
	copy(n, baseNonce)
	// 后 4 字节为大端序号
	n[8] = byte(index >> 24)
	n[9] = byte(index >> 16)
	n[10] = byte(index >> 8)
	n[11] = byte(index)
	return n
}

// chunkAAD 构造该块的附加认证数据：把块绑定到文件与序号。
// aadPrefix 应能唯一标识该文件（本项目用存储文件名）。
func chunkAAD(aadPrefix string, index int) []byte {
	return []byte(fmt.Sprintf("%s:%d", aadPrefix, index))
}

// FileChunkCount 计算明文长度对应的块数。
func FileChunkCount(plainLen int64) int {
	if plainLen <= 0 {
		return 0
	}
	return int((plainLen + FileChunkSize - 1) / FileChunkSize)
}

// FilePlainChunkLen 第 i 块的明文长度。
func FilePlainChunkLen(plainLen int64, i int) int64 {
	off := int64(i) * int64(FileChunkSize)
	if off >= plainLen {
		return 0
	}
	rest := plainLen - off
	if rest > int64(FileChunkSize) {
		return int64(FileChunkSize)
	}
	return rest
}

// FileCipherChunkLen 第 i 块的密文长度（明文长度 + tag）。
func FileCipherChunkLen(plainLen int64, i int) int64 {
	n := FilePlainChunkLen(plainLen, i)
	if n == 0 {
		return 0
	}
	return n + 16 // GCM tag
}

// FileCipherLen 整体密文长度：明文长度 + 每块 16 字节 tag。
func FileCipherLen(plainLen int64) int64 {
	return plainLen + int64(FileChunkCount(plainLen))*16
}

// FileChunkCipherOffset 第 i 块在密文中的起始偏移。
func FileChunkCipherOffset(plainLen int64, i int) int64 {
	var off int64
	for k := 0; k < i; k++ {
		off += FileCipherChunkLen(plainLen, k)
	}
	return off
}

// EncryptFileBytes 加密整份文件内容（内存方式，用于已读入内存的文件）。
func EncryptFileBytes(fek, baseNonce []byte, aadPrefix string, plain []byte) ([]byte, error) {
	if len(baseNonce) != FileNonceSize {
		return nil, fmt.Errorf("secure: nonce 长度应为 %d", FileNonceSize)
	}
	gcm, err := newGCM(fek)
	if err != nil {
		return nil, err
	}

	total := FileChunkCount(int64(len(plain)))
	out := make([]byte, 0, FileCipherLen(int64(len(plain))))

	for i := 0; i < total; i++ {
		start := i * FileChunkSize
		end := start + FileChunkSize
		if end > len(plain) {
			end = len(plain)
		}
		ct := gcm.Seal(nil, chunkNonce(baseNonce, i), plain[start:end], chunkAAD(aadPrefix, i))
		out = append(out, ct...)
	}
	return out, nil
}

// DecryptFileBytes 解密整份文件内容。
func DecryptFileBytes(fek, baseNonce []byte, aadPrefix string, cipherBytes []byte, plainLen int64) ([]byte, error) {
	if plainLen == 0 {
		return []byte{}, nil
	}
	// 落在整文件区间内即等价于解密全部
	full, err := DecryptFileRange(fek, baseNonce, aadPrefix, cipherBytes, plainLen, 0, plainLen-1)
	if err != nil {
		return nil, err
	}
	return full, nil
}

// DecryptFileRange 解密明文字节区间 [start, end]（两端含），用于 HTTP Range。
//
// 只读取并解密覆盖该区间的块，其余块完全不读、不解密。
func DecryptFileRange(fek, baseNonce []byte, aadPrefix string, cipherBytes []byte, plainLen, start, end int64) ([]byte, error) {
	if start < 0 || end < start || end >= plainLen {
		return nil, fmt.Errorf("secure: 非法区间 [%d,%d]（明文长度 %d）", start, end, plainLen)
	}
	if len(baseNonce) != FileNonceSize {
		return nil, fmt.Errorf("secure: nonce 长度应为 %d", FileNonceSize)
	}
	gcm, err := newGCM(fek)
	if err != nil {
		return nil, err
	}

	firstChunk := int(start / FileChunkSize)
	lastChunk := int(end / FileChunkSize)

	var out []byte
	for i := firstChunk; i <= lastChunk; i++ {
		cOff := FileChunkCipherOffset(plainLen, i)
		cLen := FileCipherChunkLen(plainLen, i)
		if cOff+cLen > int64(len(cipherBytes)) {
			return nil, fmt.Errorf("secure: 密文长度不足（块 %d 需要 [%d,%d)，实际 %d）", i, cOff, cOff+cLen, len(cipherBytes))
		}
		pt, err := gcm.Open(nil, chunkNonce(baseNonce, i), cipherBytes[cOff:cOff+cLen], chunkAAD(aadPrefix, i))
		if err != nil {
			return nil, fmt.Errorf("secure: 第 %d 块解密失败（密钥不符或数据被篡改）: %w", i, err)
		}

		chunkPlainStart := int64(i) * int64(FileChunkSize)
		// 该块内需要输出的区间
		lo := int64(0)
		if start > chunkPlainStart {
			lo = start - chunkPlainStart
		}
		hi := int64(len(pt))
		if end < chunkPlainStart+int64(len(pt))-1 {
			hi = end - chunkPlainStart + 1
		}
		if lo >= hi {
			continue
		}
		out = append(out, pt[lo:hi]...)
	}
	return out, nil
}
