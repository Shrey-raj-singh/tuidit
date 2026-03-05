# Homebrew formula for tuidit. To use:
#   brew tap Shrey-raj-singh/tuidit https://github.com/Shrey-raj-singh/tuidit
#   brew install tuidit
# Or: brew install Shrey-raj-singh/tuidit/tuidit
class Tuidit < Formula
  desc "Terminal UI code editor with file explorer and vim-like editing"
  homepage "https://github.com/Shrey-raj-singh/tuidit"
  version "0.1.1"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/Shrey-raj-singh/tuidit/releases/download/v0.1.1/tuidit-darwin-amd64"
      sha256 "REPLACE_WITH_DARWIN_AMD64_SHA256"
    end
    on_arm do
      url "https://github.com/Shrey-raj-singh/tuidit/releases/download/v0.1.0/tuidit-darwin-arm64"
      sha256 "REPLACE_WITH_DARWIN_ARM64_SHA256"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/Shrey-raj-singh/tuidit/releases/download/v0.1.1/tuidit-linux-amd64"
      sha256 "REPLACE_WITH_LINUX_AMD64_SHA256"
    end
    on_arm do
      url "https://github.com/Shrey-raj-singh/tuidit/releases/download/v0.1.1/tuidit-linux-arm64"
      sha256 "REPLACE_WITH_LINUX_ARM64_SHA256"
    end
  end

  def install
    if OS.mac?
      bin.install (Hardware::CPU.arm? ? "tuidit-darwin-arm64" : "tuidit-darwin-amd64") => "tuidit"
    else
      bin.install (Hardware::CPU.arm? ? "tuidit-linux-arm64" : "tuidit-linux-amd64") => "tuidit"
    end
  end

  test do
    assert_match(/tuidit .+/, shell_output("#{bin}/tuidit --version"))
  end
end
