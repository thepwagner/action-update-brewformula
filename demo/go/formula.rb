class GoLang < DebianFormula
  VERSION = '1.15.8'

  name 'golang'
  homepage 'http://www.golang.org'
  url "https://golang.org/dl/go#{VERSION}.linux-amd64.tar.gz"
  sha256 'd3379c32a90fdf9382166f8f48034c459a8cc433730bc9476d39d9082c94583b'

  version "#{VERSION}+thepwagner1"
end
