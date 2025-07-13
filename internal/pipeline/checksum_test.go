package pipeline

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChecksumWriter(t *testing.T) {
	testData := []byte("test data for checksum")
	
	// 예상 체크섬 값들
	expectedMD5 := md5.Sum(testData)
	expectedSHA256 := sha256.Sum256(testData)
	expectedSHA512 := sha512.Sum512(testData)

	t.Run("calculates MD5 checksum", func(t *testing.T) {
		cw := NewChecksumWriter("md5")
		
		n, err := cw.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		
		sum := cw.Sum()
		assert.Equal(t, hex.EncodeToString(expectedMD5[:]), sum)
	})

	t.Run("calculates SHA256 checksum", func(t *testing.T) {
		cw := NewChecksumWriter("sha256")
		
		n, err := cw.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		
		sum := cw.Sum()
		assert.Equal(t, hex.EncodeToString(expectedSHA256[:]), sum)
	})

	t.Run("calculates SHA512 checksum", func(t *testing.T) {
		cw := NewChecksumWriter("sha512")
		
		n, err := cw.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		
		sum := cw.Sum()
		assert.Equal(t, hex.EncodeToString(expectedSHA512[:]), sum)
	})

	t.Run("handles multiple writes", func(t *testing.T) {
		cw := NewChecksumWriter("sha256")
		
		// 여러 번 나누어 쓰기
		cw.Write([]byte("test "))
		cw.Write([]byte("data "))
		cw.Write([]byte("for "))
		cw.Write([]byte("checksum"))
		
		sum := cw.Sum()
		assert.Equal(t, hex.EncodeToString(expectedSHA256[:]), sum)
	})

	t.Run("supports unknown algorithm with default", func(t *testing.T) {
		cw := NewChecksumWriter("unknown")
		
		// 알 수 없는 알고리즘은 sha256으로 대체
		n, err := cw.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		
		sum := cw.Sum()
		assert.Equal(t, hex.EncodeToString(expectedSHA256[:]), sum)
	})

	t.Run("thread safe operations", func(t *testing.T) {
		cw := NewChecksumWriter("md5")
		
		// 동시에 여러 고루틴에서 쓰기
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				cw.Write([]byte("x"))
				done <- true
			}()
		}
		
		// 모든 고루틴 완료 대기
		for i := 0; i < 10; i++ {
			<-done
		}
		
		// 10개의 'x' 문자에 대한 MD5
		expected := md5.Sum([]byte("xxxxxxxxxx"))
		assert.Equal(t, hex.EncodeToString(expected[:]), cw.Sum())
	})
}

func TestChecksumFileWriter(t *testing.T) {
	tempDir := t.TempDir()
	testData := []byte("file content for checksum")

	t.Run("saves checksum to file", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "test.txt")
		cfw := NewChecksumFileWriter("sha256", outputPath)
		
		// 데이터 쓰기
		n, err := cfw.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		
		// 파일 저장
		err = cfw.SaveToFile()
		assert.NoError(t, err)
		
		// 체크섬 파일이 생성되었는지 확인
		checksumPath := outputPath + ".sha256"
		assert.FileExists(t, checksumPath)
		
		// 파일 내용 확인
		content, err := os.ReadFile(checksumPath)
		require.NoError(t, err)
		
		expected := sha256.Sum256(testData)
		expectedContent := hex.EncodeToString(expected[:]) + "  " + filepath.Base(outputPath) + "\n"
		assert.Equal(t, expectedContent, string(content))
	})

	t.Run("creates checksum for different algorithms", func(t *testing.T) {
		algorithms := []string{"md5", "sha256", "sha512"}
		
		for _, algo := range algorithms {
			outputPath := filepath.Join(tempDir, "test_"+algo+".txt")
			cfw := NewChecksumFileWriter(algo, outputPath)
			
			cfw.Write(testData)
			err := cfw.SaveToFile()
			assert.NoError(t, err)
			
			checksumPath := outputPath + "." + algo
			assert.FileExists(t, checksumPath)
		}
	})

	t.Run("handles io.Copy integration", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "copy_test.txt")
		cfw := NewChecksumFileWriter("md5", outputPath)
		
		// io.Copy를 사용한 복사
		reader := bytes.NewReader(testData)
		n, err := io.Copy(cfw, reader)
		
		assert.NoError(t, err)
		assert.Equal(t, int64(len(testData)), n)
		
		// 체크섬 파일 저장
		err = cfw.SaveToFile()
		assert.NoError(t, err)
		
		checksumPath := outputPath + ".md5"
		assert.FileExists(t, checksumPath)
	})

	t.Run("atomic file write", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "atomic_test.txt")
		cfw := NewChecksumFileWriter("sha256", outputPath)
		
		// 기존 파일 생성
		oldContent := "old checksum content"
		checksumPath := outputPath + ".sha256"
		err := os.WriteFile(checksumPath, []byte(oldContent), 0644)
		require.NoError(t, err)
		
		// 새 체크섬 쓰기
		cfw.Write(testData)
		err = cfw.SaveToFile()
		assert.NoError(t, err)
		
		// 파일이 올바르게 업데이트되었는지 확인
		content, err := os.ReadFile(checksumPath)
		require.NoError(t, err)
		assert.NotEqual(t, oldContent, string(content))
	})
}