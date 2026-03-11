# Homebrew formula for tuidit. To use:
#   brew tap Shrey-raj-singh/tuidit
#   brew install tuidit
# Or directly: brew install Shrey-raj-singh/tuidit/tuidit
# Requires: https://github.com/Shrey-raj-singh/homebrew-tuidit with this file at Formula/tuidit.rb
class Tuidit < Formula
  desc "Terminal UI code editor with file explorer and vim-like editing"
  homepage "https://github.com/Shrey-raj-singh/tuidit"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/Shrey-raj-singh/tuidit/releases/download/v0.1.0/tuidit-darwin-amd64"
      sha256 "befec55f26d63a277a05c9b948f2389bed9311cfd0bf3a045484417711b7a5b2"
    end
    on_arm do
      url "https://github.com/Shrey-raj-singh/tuidit/releases/download/v0.1.0/tuidit-darwin-arm64"
      sha256 "47b91efe3e0d369be8f91360254d2d37f6a0506973638b3ce311513c68051eb7"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/Shrey-raj-singh/tuidit/releases/download/v0.1.0/tuidit-linux-amd64"
      sha256 "eaa90a8d6ddd460d96ec697225f158af0d51f3b72c5c1c001145e84a1c10dad5"
    end
    on_arm do
      url "https://github.com/Shrey-raj-singh/tuidit/releases/download/v0.1.0/tuidit-linux-arm64"
      sha256 "206247f92cbaa4567db12ad00af40bfaf6811cad28360926dbc00aa53c1a3f4f"
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
