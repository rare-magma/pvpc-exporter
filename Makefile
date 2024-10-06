.PHONY: install
install:
	@mkdir --parents $${HOME}/.local/bin \
	&& mkdir --parents $${HOME}/.config/systemd/user \
	&& cp pvpc_exporter $${HOME}/.local/bin/ \
	&& cp --no-clobber pvpc_exporter.json $${HOME}/.config/pvpc_exporter.json \
	&& chmod 400 $${HOME}/.config/pvpc_exporter.json \
	&& cp pvpc-exporter.timer $${HOME}/.config/systemd/user/ \
	&& cp pvpc-exporter.service $${HOME}/.config/systemd/user/ \
	&& systemctl --user enable --now pvpc-exporter.timer

.PHONY: uninstall
uninstall:
	@rm -f $${HOME}/.local/bin/pvpc_exporter \
	&& rm -f $${HOME}/.config/pvpc_exporter.json \
	&& systemctl --user disable --now pvpc-exporter.timer \
	&& rm -f $${HOME}/.config/.config/systemd/user/pvpc-exporter.timer \
	&& rm -f $${HOME}/.config/systemd/user/pvpc-exporter.service

.PHONY: build
build:
	@go build -ldflags="-s -w" -o pvpc_exporter main.go