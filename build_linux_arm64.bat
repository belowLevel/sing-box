set VERSION=1.14.3

set GOOS=linux
set GOARCH=arm64
go build -v -trimpath -buildvcs=false -o dist/sing-box -tags "with_gvisor with_quic with_wireguard with_utls with_clash_api with_tailscale with_openvpn with_openconnect badlinkname tfogo_checklinkname0 with_musl" -ldflags "-X 'github.com/sagernet/sing-box/constant.Version=%VERSION%' -s -w -buildid=" ./cmd/sing-box
		
