class AzCopy < DebianFormula
  homepage 'https://github.com/Azure/azure-storage-azcopy/'
  version '10.7.0'
  url "https://github.com/Azure/azure-storage-azcopy/archive/v#{version}.tar.gz"
  sha256 'cfdc53dd2c5d30adddeb5270310ff566b4417a9f5eec6c9f6dfbe10d1feb6213'

  name 'azcopy'
  arch 'x86_64'

  build_depends 'golang (= 1.15.8+thepwagner1)'
end
