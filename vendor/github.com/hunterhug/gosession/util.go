package gosession

import (
	"fmt"
	"github.com/gofrs/uuid"
	"strconv"
	"strings"
	"time"
)

// GetGUID gen random uuid
func GetGUID() (valueGUID string) {
	objID, err := uuid.NewV4()
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	objIdStr := objID.String()
	objIdStr = strings.Replace(objIdStr, "-", "", -1)
	valueGUID = objIdStr
	return valueGUID
}

func SI(s string) (i int64) {
	i, _ = strconv.ParseInt(s, 10, 64)
	return
}
