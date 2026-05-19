class Jamshid < Formula
  desc "CLI tool for managing multiple Claude Code profiles"
  homepage "https://github.com/PapaDanielVi/jamshid"
  version "0.1.0-beta.1"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/PapaDanielVi/jamshid/releases/download/v#{version}/jamshid_Darwin_arm64.tar.gz"
      sha256 :no_check
    else
      url "https://github.com/PapaDanielVi/jamshid/releases/download/v#{version}/jamshid_Darwin_x86_64.tar.gz"
      sha256 :no_check
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/PapaDanielVi/jamshid/releases/download/v#{version}/jamshid_Linux_arm64.tar.gz"
      sha256 :no_check
    else
      url "https://github.com/PapaDanielVi/jamshid/releases/download/v#{version}/jamshid_Linux_x86_64.tar.gz"
      sha256 :no_check
    end
  end

  def install
    bin.install "jamshid"
  end

  test do
    system "#{bin}/jamshid", "--version"
  end
end
