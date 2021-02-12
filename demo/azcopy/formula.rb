class AzCopy < DebianFormula
  homepage 'https://github.com/Azure/azure-storage-azcopy/'
  version '10.8.0'
  url "https://github.com/Azure/azure-storage-azcopy/archive/v#{version}.tar.gz"
  sha256 'd5558cd419c8d46bdc958064cb97f963d1ea793866414c025906ec15033512ed'

  name 'azcopy'
  arch 'x86_64'

  build_depends 'golang (= 1.15.6+thepwagner1)'
end
