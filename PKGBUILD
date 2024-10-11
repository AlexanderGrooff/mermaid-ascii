pkgname='mermaid-ascii'
pkgver=0.3.5
pkgrel=1
pkgdesc='Render Mermaid ASCII in textual form'
arch=(x86_64)
url="https://github.com/AlexanderGrooff/${pkgname}"
license=('MIT')
makedepends=('go')
source=("git+${url}.git#tag=${pkgver}")
sha256sums=('SKIP')

pkgver() {
  cd "${pkgname}"
  git describe --tags
}

prepare(){
  cd "${pkgname}"
  mkdir -p build/
}

build() {
  cd "${pkgname}"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
  go build -o build
}

check() {
  cd "${pkgname}"
  go test ./...
}

package() {
  cd "${pkgname}"
  install -Dm755 build/${pkgname} "$pkgdir"/usr/bin/${pkgname}
}
