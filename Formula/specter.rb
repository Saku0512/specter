class Specter < Formula
  desc "Lightweight mock API server"
  homepage "https://github.com/Saku0512/specter"
  version "1.0.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_darwin_arm64"
      sha256 "cdeb04c2393e1d6e7f765b09028ce00940eed34bfe4b546730845db78198c1e4"
    end
    on_intel do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_darwin_amd64"
      sha256 "71f610799f24d1cc53942f1238a109315d88d83a774d5272c7c3f2963960c778"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_linux_arm64"
      sha256 "91c39662bec08f10b9b4b6f153e24067c875eef7c91cb10ff53a898e36ebb4c5"
    end
    on_intel do
      url "https://github.com/Saku0512/specter/releases/download/v#{version}/specter_linux_amd64"
      sha256 "099e341bd641e9fadd4d8e710af63e028edbbe4f1bb42b9e171edfd3de1b85fb"
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
