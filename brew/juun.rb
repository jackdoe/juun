class Juun < Formula
  desc "juun - bash/zsh history manager and search"
  homepage "https://github.com/jackdoe/juun"
  url "https://github.com/jackdoe/juun/archive/ff9f22f851bd2de2ce375da20792b1856d5e74a0.zip"
  version "0.1"
  sha256 "219a45a825a3fd0d669156fe775668ae22428e05"
  depends_on "golang" => :build, "make" => :build
  def install
    system "make", "install" 
  end

  test do
    system "true"
  end
end

