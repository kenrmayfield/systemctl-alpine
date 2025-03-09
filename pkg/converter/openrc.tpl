#!/sbin/openrc-run

name="{{.Name}}"
description="{{.Description}}"
{{if .User}}
command_user="{{.User}}{{if .Group}}:{{.Group}}{{end}}"
{{end}}
{{if .WorkingDirectory}}
directory="{{.WorkingDirectory}}"
{{end}}
command_background="yes"

command="{{.Command}}"
{{if .CommandArgs}}
command_args="{{.CommandArgs}}"
{{end}}

pidfile="/run/$name/$name.pid"

{{if .Capabilities}}
capabilities="{{.Capabilities}}"
{{end}}

{{if .EnvironmentFile}}
# Source environment file if it exists
if [ -f "{{.EnvironmentFile}}" ]; then
	export $(grep -v '^#' "{{.EnvironmentFile}}" | xargs)
fi
{{end}}

{{range .Environment}}
export {{.}}
{{end}}

depend() {
    need net
#    after firewall
}

start_pre() {
    checkpath --directory --owner $command_user --mode 0755 ${pidfile%/*}
{{range .ExecStartPreCommands}}
    {{.}}
{{end}}
}

{{if .StopCommand}}
stop() {
    ebegin "Stopping $RC_SVCNAME"
    {{.StopCommand}}
    eend $?
}
{{end}}

reload() {
    ebegin "Reloading $RC_SVCNAME configuration"
    start_pre && start-stop-daemon --signal HUP --pidfile $pidfile
    eend $?
}
