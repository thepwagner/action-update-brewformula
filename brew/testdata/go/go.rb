class GoLang < DebianFormula
  VERSION = '1.15.6'

  name 'golang'
  homepage 'http://www.golang.org'
  url "https://dl.google.com/go/go#{VERSION}.linux-amd64.tar.gz"
  sha256 '3918e6cc85e7eaaa6f859f1bdbaac772e7a825b0eb423c63d3ae68b21f84b844'

  version "#{VERSION}+thepwagner1"
end
