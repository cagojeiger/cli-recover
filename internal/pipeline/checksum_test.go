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

func TestChecksumWriterMonitorInterface(t *testing.T) {
	t.Run("implements Monitor interface", func(t *testing.T) {
		cw := NewChecksumWriter("sha256")
		
		// Monitor 인터페이스 구현 확인
		var _ Monitor = cw
		
		// Update 메서드
		cw.Update(100)
		// Update는 ChecksumWriter에서 무시됨
		
		// Finish 메서드
		cw.Finish()
		// Finish는 ChecksumWriter에서 특별한 동작 없음
		
		// Report 메서드
		testData := []byte("test data")
		cw.Write(testData)
		
		report := cw.Report()
		expected := sha256.Sum256(testData)
		expectedReport := "Checksum (sha256): " + hex.EncodeToString(expected[:])
		assert.Equal(t, expectedReport, report)
	})
	
	t.Run("Algorithm method", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"md5", "md5"},
			{"sha256", "sha256"},
			{"sha512", "sha512"},
			{"unknown", "unknown"},
		}
		
		for _, tc := range tests {
			cw := NewChecksumWriter(tc.input)
			assert.Equal(t, tc.expected, cw.Algorithm())
		}
	})
}

func TestMultiChecksumWriter(t *testing.T) {
	testData := []byte("multi checksum test data")
	
	t.Run("calculates multiple checksums", func(t *testing.T) {
		algorithms := []string{"md5", "sha256", "sha512"}
		mcw := NewMultiChecksumWriter(algorithms)
		
		// 데이터 쓰기
		n, err := mcw.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		
		// 모든 체크섬 가져오기
		sums := mcw.Sums()
		assert.Len(t, sums, 3)
		
		// 각 체크섬 검증
		expectedMD5 := md5.Sum(testData)
		assert.Equal(t, hex.EncodeToString(expectedMD5[:]), sums["md5"])
		
		expectedSHA256 := sha256.Sum256(testData)
		assert.Equal(t, hex.EncodeToString(expectedSHA256[:]), sums["sha256"])
		
		expectedSHA512 := sha512.Sum512(testData)
		assert.Equal(t, hex.EncodeToString(expectedSHA512[:]), sums["sha512"])
	})
	
	t.Run("handles empty algorithms list", func(t *testing.T) {
		mcw := NewMultiChecksumWriter([]string{})
		
		n, err := mcw.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		
		sums := mcw.Sums()
		assert.Empty(t, sums)
	})
	
	t.Run("thread safe writes", func(t *testing.T) {
		mcw := NewMultiChecksumWriter([]string{"sha256"})
		
		// 동시에 여러 고루틴에서 쓰기
		done := make(chan bool, 5)
		chunk := []byte("chunk")
		
		for i := 0; i < 5; i++ {
			go func() {
				mcw.Write(chunk)
				done <- true
			}()
		}
		
		// 모든 고루틴 완료 대기
		for i := 0; i < 5; i++ {
			<-done
		}
		
		sums := mcw.Sums()
		// 5개의 "chunk" = "chunkchunkchunkchunkchunk"
		expected := sha256.Sum256([]byte("chunkchunkchunkchunkchunk"))
		assert.Equal(t, hex.EncodeToString(expected[:]), sums["sha256"])
	})
}

func TestChecksumVerifier(t *testing.T) {
	tempDir := t.TempDir()
	
	t.Run("verifies valid checksum", func(t *testing.T) {
		// 테스트 파일 생성
		testFile := filepath.Join(tempDir, "test.txt")
		testData := []byte("verification test data")
		err := os.WriteFile(testFile, testData, 0644)
		require.NoError(t, err)
		
		// 체크섬 파일 생성
		expected := sha256.Sum256(testData)
		checksumContent := hex.EncodeToString(expected[:]) + "  test.txt\n"
		checksumFile := testFile + ".sha256"
		err = os.WriteFile(checksumFile, []byte(checksumContent), 0644)
		require.NoError(t, err)
		
		// 검증
		verifier := NewChecksumVerifier("sha256")
		valid, err := verifier.VerifyFile(testFile, hex.EncodeToString(expected[:]))
		assert.NoError(t, err)
		assert.True(t, valid)
	})
	
	t.Run("detects invalid checksum", func(t *testing.T) {
		// 테스트 파일 생성
		testFile := filepath.Join(tempDir, "invalid.txt")
		testData := []byte("invalid verification test")
		err := os.WriteFile(testFile, testData, 0644)
		require.NoError(t, err)
		
		// 잘못된 체크섬 파일 생성
		checksumContent := "invalidchecksum1234567890abcdef1234567890abcdef1234567890abcdef12  invalid.txt\n"
		checksumFile := testFile + ".sha256"
		err = os.WriteFile(checksumFile, []byte(checksumContent), 0644)
		require.NoError(t, err)
		
		// 검증
		verifier := NewChecksumVerifier("sha256")
		valid, err := verifier.VerifyFile(testFile, "invalidchecksum1234567890abcdef1234567890abcdef1234567890abcdef12")
		assert.NoError(t, err)
		assert.False(t, valid)
	})
	
	t.Run("handles missing checksum file", func(t *testing.T) {
		// 체크섬 파일이 없는 파일
		testFile := filepath.Join(tempDir, "no-checksum.txt")
		err := os.WriteFile(testFile, []byte("no checksum"), 0644)
		require.NoError(t, err)
		
		verifier := NewChecksumVerifier("sha256")
		// 체크섬 검증은 체크섬 값을 제공해야 함
		expected := sha256.Sum256([]byte("no checksum"))
		valid, err := verifier.VerifyFile(testFile, hex.EncodeToString(expected[:]))
		assert.NoError(t, err)
		assert.True(t, valid)
	})
	
	t.Run("handles missing target file", func(t *testing.T) {
		// 존재하지 않는 파일
		testFile := filepath.Join(tempDir, "nonexistent.txt")
		
		verifier := NewChecksumVerifier("sha256")
		valid, err := verifier.VerifyFile(testFile, "dummychecksum")
		assert.Error(t, err)
		assert.False(t, valid)
	})
	
	t.Run("supports different algorithms", func(t *testing.T) {
		algorithms := []string{"md5", "sha256", "sha512"}
		
		for _, algo := range algorithms {
			// 각 알고리즘에 대한 verifier 생성
			verifier := NewChecksumVerifier(algo)
			assert.NotNil(t, verifier)
			
			// 실제 파일로 테스트
			testFile := filepath.Join(tempDir, "algo-"+algo+".txt")
			testData := []byte("algorithm test: " + algo)
			err := os.WriteFile(testFile, testData, 0644)
			require.NoError(t, err)
			
			// 체크섬 계산
			var checksumHex string
			switch algo {
			case "md5":
				sum := md5.Sum(testData)
				checksumHex = hex.EncodeToString(sum[:])
			case "sha256":
				sum := sha256.Sum256(testData)
				checksumHex = hex.EncodeToString(sum[:])
			case "sha512":
				sum := sha512.Sum512(testData)
				checksumHex = hex.EncodeToString(sum[:])
			}
			
			// 체크섬 파일 생성
			checksumContent := checksumHex + "  " + filepath.Base(testFile) + "\n"
			checksumFile := testFile + "." + algo
			err = os.WriteFile(checksumFile, []byte(checksumContent), 0644)
			require.NoError(t, err)
			
			// 검증
			valid, err := verifier.VerifyFile(testFile, checksumHex)
			assert.NoError(t, err)
			assert.True(t, valid)
		}
	})
}