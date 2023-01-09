.PHONY: install
install:
	@mkdir --parents $${HOME}/.local/bin \
	&& mkdir --parents $${HOME}/.config/systemd/user \
	&& cp pvpc_exporter.sh $${HOME}/.local/bin/ \
	&& chmod +x $${HOME}/.local/bin/pvpc_exporter.sh \
	&& cp --no-clobber pvpc_exporter.conf $${HOME}/.config/pvpc_exporter.conf \
	&& chmod 400 $${HOME}/.config/pbs_exporter.conf \
	&& cp pvpc-exporter.timer $${HOME}/.config/systemd/user/ \
	&& cp pvpc-exporter.service $${HOME}/.config/systemd/user/ \
	&& systemctl --user enable --now pvpc-exporter.timer

.PHONY: uninstall
uninstall:
	@rm -f $${HOME}/.local/bin/pvpc_exporter.sh \
	&& rm -f $${HOME}/.config/pvpc_exporter.conf \
	&& systemctl --user disable --now pvpc-exporter.timer \
	&& rm -f $${HOME}/.config/.config/systemd/user/pvpc-exporter.timer \
	&& rm -f $${HOME}/.config/systemd/user/pvpc-exporter.service
