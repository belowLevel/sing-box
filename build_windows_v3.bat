set GOARCH=amd64
set GOAMD64=v3

go build -v -trimpath -buildvcs=false -o dist/sing-box.exe -tags "with_gvisor with_quic with_dhcp with_wireguard with_utls with_acme with_clash_api with_tailscale with_ccm with_ocm with_cloudflared with_usbip with_openvpn with_openconnect badlinkname with_musl" -ldflags "-s -w -buildid=" ./cmd/sing-box
		