# https://github.com/tmm1/brew2deb/tree/ad77608d0dd1d6a7d72e5514ba0d86f4de381288/packages/libvirt

class LibVirt < DebianFormula
  homepage 'http://www.libvirt.org'
  url 'https://libvirt.org/sources/libvirt-1.0.2.tar.gz'
  sha256 '9b8c2752f78658b65ef1c608b3775be0978d60855a9b5e2778f79c113201c179'


  name 'libvirt'
  section 'libs'
  version '1.0.2+github1'
  description 'libvirt is a library to interact with various virtualization technologies'

  conflicts 'libvirt-bin'
  replaces  'libvirt-bin'
  provides  'libvirt-bin'

  build_depends \
    'libxml2-dev',
    'libgnutls-dev',
    'libdevmapper-dev',
    'libcurl4-gnutls-dev',
    'python-dev',
    'libnl-dev',
    'libudev-dev',
    'libpciaccess-dev'

  depends \
    'libxml2',
    'libgnutls26',
    'libdevmapper',
    'libcurl3-gnutls',
    'libnl1',
    'udev',
    'libpciaccess0'

  def patches
    [ 'prevent-jvm-segfault.patch',
      'esx-thin-provision.patch']
  end

  def build
  end

  def install
    args = ["--prefix=#{prefix}",
            "--with-esx",
            "--with-remote",
            "--with-test",
            "--with-vbox",
            "--with-vmware",
            "--with-kvm",
            "--without-qemu",
            "--without-xen",
           ]

    sh "./configure", *args

    # Compilation of docs doesn't get done if we jump straight to "make install"
    sh "make"
    sh "make install"

  end

end