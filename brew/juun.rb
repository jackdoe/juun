class Juun < Formula
  desc "juun - bash/zsh history manager and search"
  homepage "https://github.com/jackdoe/juun"
  url "https://github.com/jackdoe/juun/archive/d1975731f624cd902e0f188cf8e4a5343fd65308.zip"
  version "0.1"
  sha256 "3e626f9d0a5fbd026e300c3965f83df4aa09095f"
  depends_on "golang" => :build, "make" => :build
  def install
    system "make", "install" 
  end

  test do
    system "true"
  end
end

