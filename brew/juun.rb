class Juun < Formula
  desc "juun - bash/zsh history manager and search"
  homepage "https://github.com/jackdoe/juun"
  url "https://github.com/jackdoe/juun/archive/f98dee2b47b7c31cd48669a285ae0169acbea04e.zip"
  version "0.2"
  sha256 "b04d052fc8744b9fe3ea71c400700aa5cbedeb52"
  depends_on "golang" => :build, "make" => :build
  def install
    system "make", "install" 
  end

  test do
    system "true"
  end
end

