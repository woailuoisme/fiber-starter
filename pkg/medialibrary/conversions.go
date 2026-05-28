package medialibrary

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	chaiwebp "github.com/chai2010/webp"
	"github.com/disintegration/gift"
	xwebp "golang.org/x/image/webp"
)

// 自动注册 Go 图像标准库的解码器，支持常用图片读取
func init() {
	image.RegisterFormat("jpeg", "\xff\xd8\xff", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "\x89PNG\r\n\x1a\n", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "GIF8", gif.Decode, gif.DecodeConfig)
	image.RegisterFormat("webp", "RIFF????WEBP", xwebp.Decode, xwebp.DecodeConfig)
}

// performImageConversion 接收源图像字节、转换规则定义，执行图像处理并返回处理后的二进制数据
func performImageConversion(srcData []byte, rule *ConversionRule) ([]byte, string, error) {
	// 1. 解码源图像
	srcImg, format, err := image.Decode(bytes.NewReader(srcData))
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode source image: %w", err)
	}

	// 2. 规范化目标图片格式定义
	targetFormat := strings.ToLower(rule.Format)
	if targetFormat == "" || targetFormat == "jpg" {
		targetFormat = format
	}

	// 3. 构建 GIFT 图像滤镜处理器
	g := gift.New()

	if rule.Crop && rule.Width > 0 && rule.Height > 0 {
		// 居中裁剪缩放：比例自适应填满目标，然后按中心裁剪到精确的宽高
		g.Add(gift.ResizeToFill(rule.Width, rule.Height, gift.LanczosResampling, gift.CenterAnchor))
	} else {
		// 等比例自适应缩放
		switch {
		case rule.Width > 0 && rule.Height > 0:
			g.Add(gift.ResizeToFit(rule.Width, rule.Height, gift.LanczosResampling))
		case rule.Width > 0:
			g.Add(gift.Resize(rule.Width, 0, gift.LanczosResampling))
		case rule.Height > 0:
			g.Add(gift.Resize(0, rule.Height, gift.LanczosResampling))
		}
	}

	// 4. 运行图像处理
	dstImg := image.NewRGBA(g.Bounds(srcImg.Bounds()))
	g.Draw(dstImg, srcImg)

	// 5. 编码到目标格式二进制数据中
	var outBuf bytes.Buffer
	err = encodeImage(&outBuf, dstImg, targetFormat)
	if err != nil {
		return nil, "", err
	}

	return outBuf.Bytes(), targetFormat, nil
}

// encodeImage 辅助编码器，支持 JPEG, PNG, GIF
func encodeImage(w io.Writer, img image.Image, format string) error {
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 85})
	case "png":
		return png.Encode(w, img)
	case "gif":
		return gif.Encode(w, img, nil)
	case "webp":
		return chaiwebp.Encode(w, img, &chaiwebp.Options{Lossless: false, Quality: 80})
	default:
		// 默认兜底使用 PNG 无损编码
		return png.Encode(w, img)
	}
}
