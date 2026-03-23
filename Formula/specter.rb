class Specter < Formula
  desc "Lightweight mock API server"
  homepage "https://github.com/Saku0512/specter"
  version "0.3.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_darwin_arm64"
      sha256 "c61b7fd41c520ba20eb50527e226bc425c22e5b70a4e1585c956b446db116f39"
    end
    on_intel do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_darwin_amd64"
      sha256 "0d7503574d8b1c847fcb879799289be4d2fe509d33487fccd7e229a6e4989afc"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_linux_arm64"
      sha256 "5b98445bdee680e354953b941b90159e4b384a7fc97a4c1f78484d4e2dc4f64c"
    end
    on_intel do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_linux_amd64"
      sha256 "31ad419ac80ff51553901f377db57e05dcd18eb7d1815fccf43a1c0ddbaea1e8"
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
