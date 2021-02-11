class AzCopy < DebianFormula
  homepage 'https://github.com/Azure/azure-storage-azcopy/'
  version '10.8.0'
  url "https://github.com/Azure/azure-storage-azcopy/archive/v#{version}.tar.gz"
  sha256 '95866844ff1bb315879b2f1ef70f7076a4cae2391d289af474d75ee2ca3b023c'

  name 'azcopy'
  arch 'x86_64'

  build_depends 'golang (= 1.15.6+thepwagner1)'
end
