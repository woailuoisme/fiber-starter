package services

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

func (s *StorageService) initZstd() error {
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return fmt.Errorf("failed to create zstd encoder: %w", err)
	}

	decoder, err := zstd.NewReader(nil)
	if err != nil {
		_ = encoder.Close()
		return fmt.Errorf("failed to create zstd decoder: %w", err)
	}

	s.closeCompression()
	s.zstdEncoder = encoder
	s.zstdDecoder = decoder
	return nil
}

func (s *StorageService) closeCompression() {
	if s.zstdEncoder != nil {
		_ = s.zstdEncoder.Close()
		s.zstdEncoder = nil
	}
	if s.zstdDecoder != nil {
		s.zstdDecoder.Close()
		s.zstdDecoder = nil
	}
}

func (s *StorageService) compressData(data []byte) ([]byte, error) {
	switch s.compression {
	case CompressionNone:
		return data, nil
	case CompressionGzip:
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		if _, err := writer.Write(data); err != nil {
			return nil, fmt.Errorf("gzip compression failed: %w", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip writer: %w", err)
		}
		return buf.Bytes(), nil
	case CompressionZstd:
		if s.zstdEncoder == nil {
			return nil, fmt.Errorf("zstd encoder not initialized")
		}
		return s.zstdEncoder.EncodeAll(data, nil), nil
	default:
		return nil, fmt.Errorf("unsupported compression type: %d", s.compression)
	}
}

func (s *StorageService) decompressData(data []byte) ([]byte, error) {
	switch s.compression {
	case CompressionNone:
		return data, nil
	case CompressionGzip:
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer func() { _ = reader.Close() }()

		result, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("gzip decompression failed: %w", err)
		}
		return result, nil
	case CompressionZstd:
		if s.zstdDecoder == nil {
			return nil, fmt.Errorf("zstd decoder not initialized")
		}
		return s.zstdDecoder.DecodeAll(data, nil)
	default:
		return nil, fmt.Errorf("unsupported compression type: %d", s.compression)
	}
}
