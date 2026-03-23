class Specter < Formula
  desc "Lightweight mock API server"
  homepage "https://github.com/Saku0512/specter"
  version "0.0.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_darwin_arm64"
      sha256 "PLACEHOLDER_DARWIN_ARM64"
    end
    on_intel do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_darwin_amd64"
      sha256 "PLACEHOLDER_DARWIN_AMD64"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_linux_arm64"
      sha256 "PLACEHOLDER_LINUX_ARM64"
    end
    on_intel do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_linux_amd64"
      sha256 "PLACEHOLDER_LINUX_AMD64"
    end
  end

  def install
    os   = OS.mac? ? "darwin" : "linux"
    arch = Hardware::CPU.arm? ? "arm64" : "amd64"
    bin.install "specter_#{os}_#{arch}" => "specter"
  end

  test do
    system "#{bin}/specter", "--version"
  end
end
