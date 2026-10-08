package secure

import (
	"bytes"
	"strings"
	"testing"
)

// newFileKeys 生成测试用 FEK 与 nonce。
func newFileKeys(t *testing.T) (fek, nonce []byte) {
	t.Helper()
	var err error
	if fek, err = NewKey(); err != nil {
		t.Fatal(err)
	}
	if nonce, err = NewFileNonce(); err != nil {
		t.Fatal(err)
	}
	return
}

func TestFileEncryptDecrypt_RoundTrip(t *testing.T) {
	fek, nonce := newFileKeys(t)

	for _, size := range []int{
		1,
		1024,
		FileChunkSize - 1,
		FileChunkSize,     // 正好一块
		FileChunkSize + 1, // 跨两块
		3*FileChunkSize + 7,
	} {
		plain := bytes.Repeat([]byte{'x'}, size)
		if size > 100 {
			for i := range plain {
				plain[i] = byte(i % 251)
			}
		}

		ct, err := EncryptFileBytes(fek, nonce, "storage/u/other/f.bin", plain)
		if err != nil {
			t.Fatalf("size=%d 加密失败: %v", size, err)
		}

		// 密文长度 = 明文 + 每块 16 字节 tag
		wantLen := FileCipherLen(int64(size))
		if int64(len(ct)) != wantLen {
			t.Errorf("size=%d 密文长度应为 %d，实际 %d", size, wantLen, len(ct))
		}
		// 密文不应等于明文
		if bytes.Equal(ct, plain) {
			t.Errorf("size=%d 密文与明文相同", size)
		}

		got, err := DecryptFileBytes(fek, nonce, "storage/u/other/f.bin", ct, int64(size))
		if err != nil {
			t.Fatalf("size=%d 解密失败: %v", size, err)
		}
		if !bytes.Equal(got, plain) {
			t.Errorf("size=%d 往返不一致", size)
		}
	}
}

func TestFileEncrypt_EmptyFile(t *testing.T) {
	fek, nonce := newFileKeys(t)

	ct, err := EncryptFileBytes(fek, nonce, "p", nil)
	if err != nil {
		t.Fatalf("空文件加密失败: %v", err)
	}
	if len(ct) != 0 {
		t.Errorf("空文件密文应为空，实际 %d 字节", len(ct))
	}
	got, err := DecryptFileBytes(fek, nonce, "p", ct, 0)
	if err != nil {
		t.Fatalf("空文件解密失败: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("空文件解密结果应为空，实际 %d 字节", len(got))
	}
}

func TestFileDecrypt_WrongKeyOrAADFails(t *testing.T) {
	fek, nonce := newFileKeys(t)
	// 跨多块，便于验证"中间某块被篡改也能被发现"
	plain := bytes.Repeat([]byte("secret-data"), FileChunkSize/11+1000)

	ct, err := EncryptFileBytes(fek, nonce, "fileA", plain)
	if err != nil {
		t.Fatal(err)
	}

	// 错误密钥
	otherFek, _ := NewKey()
	if _, err := DecryptFileBytes(otherFek, nonce, "fileA", ct, int64(len(plain))); err == nil {
		t.Error("错误密钥解密应当失败")
	}

	// 错误 AAD（把密文当成另一个文件的）
	if _, err := DecryptFileBytes(fek, nonce, "fileB", ct, int64(len(plain))); err == nil {
		t.Error("AAD 不符解密应当失败")
	}

	// 篡改第二块中的某个字节
	tampered := append([]byte(nil), ct...)
	idx := FileChunkCipherOffset(int64(len(plain)), 1) + 8
	tampered[idx] ^= 0x01
	if _, err := DecryptFileBytes(fek, nonce, "fileA", tampered, int64(len(plain))); err == nil {
		t.Error("密文被篡改后解密应当失败")
	}
}

func TestFileDecrypt_SameePlaintextDiffersPerNonce(t *testing.T) {
	fek, nonce1 := newFileKeys(t)
	_, nonce2 := newFileKeys(t)
	plain := []byte("same-content-for-file")

	ct1, _ := EncryptFileBytes(fek, nonce1, "p", plain)
	ct2, _ := EncryptFileBytes(fek, nonce2, "p", plain)
	if bytes.Equal(ct1, ct2) {
		t.Error("不同 nonce 加密同一内容应得到不同密文")
	}
}

// ---------------------------------------------------------------------------
// Range：只解密覆盖区间的块
// ---------------------------------------------------------------------------

func TestDecryptFileRange(t *testing.T) {
	fek, nonce := newFileKeys(t)
	aad := "storage/u/other/big.bin"

	// 构造跨多块的内容，内容可校验
	size := 3*FileChunkSize + 12345
	plain := make([]byte, size)
	for i := range plain {
		plain[i] = byte((i*31 + 7) % 256)
	}

	ct, err := EncryptFileBytes(fek, nonce, aad, plain)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name       string
		start, end int64
	}{
		{"文件头", 0, 99},
		{"单个块内", 1000, 2000},
		{"跨块边界", FileChunkSize - 50, FileChunkSize + 50},
		{"正好一整块", FileChunkSize, 2*FileChunkSize - 1},
		{"最后一块", int64(size - 100), int64(size - 1)},
		{"最后一块尾部", int64(size - 1), int64(size - 1)},
		{"首字节", 0, 0},
	}

	for _, c := range cases {
		got, err := DecryptFileRange(fek, nonce, aad, ct, int64(size), c.start, c.end)
		if err != nil {
			t.Errorf("%s: 解密失败: %v", c.name, err)
			continue
		}
		want := plain[c.start : c.end+1]
		if !bytes.Equal(got, want) {
			t.Errorf("%s: 区间 [%d,%d] 内容不符（got %d 字节, want %d 字节）",
				c.name, c.start, c.end, len(got), len(want))
		}
	}

	// 非法区间
	for _, bad := range [][2]int64{{-1, 10}, {10, 5}, {0, int64(size)}} {
		if _, err := DecryptFileRange(fek, nonce, aad, ct, int64(size), bad[0], bad[1]); err == nil {
			t.Errorf("非法区间 [%d,%d] 应当报错", bad[0], bad[1])
		}
	}
}

func TestFileChunkGeometry(t *testing.T) {
	// 正好整数块
	if n := FileChunkCount(FileChunkSize); n != 1 {
		t.Errorf("1 块文件应为 1 块，实际 %d", n)
	}
	if n := FileChunkCount(FileChunkSize + 1); n != 2 {
		t.Errorf("1 块 + 1 字节应为 2 块，实际 %d", n)
	}
	if n := FileChunkCount(0); n != 0 {
		t.Errorf("空文件应为 0 块，实际 %d", n)
	}

	// 偏移与长度自洽
	plainLen := int64(2*FileChunkSize + 100)
	var off int64
	for i := 0; i < FileChunkCount(plainLen); i++ {
		if got := FileChunkCipherOffset(plainLen, i); got != off {
			t.Errorf("第 %d 块偏移应为 %d，实际 %d", i, off, got)
		}
		off += FileCipherChunkLen(plainLen, i)
	}
	if off != FileCipherLen(plainLen) {
		t.Errorf("总密文长度应为 %d，实际 %d", FileCipherLen(plainLen), off)
	}
}

func TestFileNonceHexRoundTrip(t *testing.T) {
	_, nonce := newFileKeys(t)

	h := FileNonceHex(nonce)
	if len(h) != FileNonceSize*2 {
		t.Errorf("hex 长度应为 %d，实际 %d", FileNonceSize*2, len(h))
	}
	back, err := ParseFileNonce(h)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if !bytes.Equal(back, nonce) {
		t.Error("往返不一致")
	}

	if _, err := ParseFileNonce("zz"); err == nil {
		t.Error("非法 hex 应报错")
	}
	if _, err := ParseFileNonce(strings.Repeat("ab", 4)); err == nil {
		t.Error("长度不符应报错")
	}
}
