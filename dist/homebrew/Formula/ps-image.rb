# typed: false
# frozen_string_literal: true

class PsImage < Formula
  desc "Image processing without dependencies - resize, convert, extract metadata"
  homepage "https://github.com/Barneyjm/portable-skills"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Barneyjm/portable-skills/releases/download/v#{version}/ps-image-darwin-arm64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    else
      url "https://github.com/Barneyjm/portable-skills/releases/download/v#{version}/ps-image-darwin-amd64"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Barneyjm/portable-skills/releases/download/v#{version}/ps-image-linux-arm64"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    else
      url "https://github.com/Barneyjm/portable-skills/releases/download/v#{version}/ps-image-linux-amd64"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end
  end

  def install
    binary_name = Dir["ps-image-*"].first
    bin.install binary_name => "ps-image"
  end

  def caveats
    <<~EOS
      ps-image is a self-contained binary with no dependencies.

      Usage:
        ps-image resize <input> --width 800
        ps-image convert <input> --format png
        ps-image info <input> --json

      To install as a Claude Code skill:
        ps-image install-skill

      For more information:
        ps-image --help
    EOS
  end

  test do
    # Create a minimal valid PNG file (1x1 red pixel)
    (testpath/"test.png").write([
      0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,  # PNG signature
      0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,  # IHDR chunk
      0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,  # 1x1 dimensions
      0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,  # 8-bit RGB
      0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,  # IDAT chunk
      0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,  # compressed data
      0x00, 0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xFE,
      0xD4, 0xEF, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,  # IEND chunk
      0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82
    ].pack("C*"))

    # Test info command
    output = shell_output("#{bin}/ps-image info #{testpath}/test.png --json")
    assert_match '"width": 1', output
    assert_match '"height": 1', output
    assert_match '"format": "png"', output

    # Test version
    assert_match version.to_s, shell_output("#{bin}/ps-image --version")
  end
end
