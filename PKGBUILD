pkgname=codecrafters-cli
pkgver=34
pkgrel=1
pkgdesc="CodeCrafters CLI - a tool for codecrafters.io exercises"
arch=('x86_64' 'aarch64')
url="https://github.com/codecrafters-io/cli"
license=('MIT') # Adjust according to the project's actual license
depends=('curl')

_arch=$(uname -m)
if [ "$_arch" == "x86_64" ]; then
	_arch="amd64"
elif [[ "$_arch" =~ ^(armv8|arm64|aarch64)$ ]]; then
	_arch="arm64"
fi

source=("${pkgname}-${pkgver}.tar.gz::https://github.com/codecrafters-io/cli/releases/download/v${pkgver}/v${pkgver}_linux_${_arch}.tar.gz")
sha256sums=('SKIP') # Replace SKIP with the actual checksum if available

prepare() {
	echo "Preparing ${pkgname} package"
}

build() {
	echo "Building ${pkgname} package"
}

package() {
	# Create installation directory
	install -Dm755 "$srcdir/codecrafters" "$pkgdir/usr/local/bin/codecrafters"

	# Set executable permissions for codecrafters binary
	chmod 0755 "$pkgdir/usr/local/bin/codecrafters"
}

post_install() {
	echo "CodeCrafters CLI installed successfully!"
}
