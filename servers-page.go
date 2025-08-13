package channelz

import (
	"context"
	"io"

	channelzgrpc "google.golang.org/grpc/channelz/grpc_channelz_v1"
	log "google.golang.org/grpc/grpclog"
)

type Server struct {
	Server  *channelzgrpc.Server
	Sockets *channelzgrpc.GetServerSocketsResponse
}

// writeServers writes HTML to w containing RPC servers stats.
//
// It includes neither a header nor footer, so you can embed this data in other pages.
func (h *grpcChannelzHandler) writeServers(w io.Writer) {
	serversRsp := h.getServers()
	servers := make(map[int64]Server, len(serversRsp.Server))
	for _, s := range serversRsp.Server {
		servers[s.Ref.ServerId] = Server{
			Server:  s,
			Sockets: h.getServerSockets(s.Ref.ServerId),
		}
	}
	if err := serversTemplate.Execute(w, servers); err != nil {
		log.Errorf("channelz: executing template: %v", err)
	}
}

func (h *grpcChannelzHandler) getServers() *channelzgrpc.GetServersResponse {
	client, err := h.connect()
	if err != nil {
		log.Errorf("Error creating channelz client %+v", err)
		return nil
	}
	ctx := context.Background()
	servers, err := client.GetServers(ctx, &channelzgrpc.GetServersRequest{})
	if err != nil {
		log.Errorf("Error querying GetServers %+v", err)
		return nil
	}
	return servers
}

func (h *grpcChannelzHandler) getServerSockets(serverID int64) *channelzgrpc.GetServerSocketsResponse {
	client, err := h.connect()
	if err != nil {
		log.Errorf("Error creating channelz client %+v", err)
		return nil
	}
	ctx := context.Background()
	sockets, err := client.GetServerSockets(ctx, &channelzgrpc.GetServerSocketsRequest{ServerId: serverID})
	if err != nil {
		log.Errorf("Error querying GetServerSockets for server %d: %+v", serverID, err)
		return nil
	}
	return sockets
}

const serversTemplateHTML = `
{{define "server-header"}}
    <tr classs="header">
        <th>Server</th>
		<th>CreationTimestamp</th>
        <th>CallsStarted</th>
        <th>CallsSucceeded</th>
        <th>CallsFailed</th>
        <th>LastCallStartedTimestamp</th>
		<th>Sockets</th>
    </tr>
{{end}}

{{define "server-body"}}
    <tr>
        <td><a href="{{link "server" .Server.Ref.ServerId}}"><b>{{.Server.Ref.ServerId}}</b> {{.Server.Ref.Name}}</a></td>
        <td>{{with .Server.Data.Trace}} {{.CreationTimestamp | timestamp}} {{end}}</td>
        <td>{{.Server.Data.CallsStarted}}</td>
        <td>{{.Server.Data.CallsSucceeded}}</td>
        <td>{{.Server.Data.CallsFailed}}</td>
        <td>{{.Server.Data.LastCallStartedTimestamp | timestamp}}</td>
		<td>
			{{range .Sockets.SocketRef }}
				<a href="{{link "socket" .SocketId}}"><b>{{.SocketId}}</b> {{.Name}}</a> <br/>
			{{end}}
		</td>
	</tr>
	{{with .Server.Data.Trace}}
		<tr classs="header">
			<th colspan=100>Events</th>
		</tr>
		<tr>
			<td>&nbsp;</td>
			<td colspan=100>
				<pre>
				{{- range .Events}}
{{.Severity}} [{{.Timestamp | timestamp}}]: {{.Description}}
				{{- end -}}
				</pre>
			</td>
		</tr>
	{{end}}
{{end}}

<p><table class="section-header" width=100%><tr align=center><td>Servers</td></tr></table></p>
<table frame=box cellspacing=0 cellpadding=2>
    <tr class="header">
		<th colspan=100 style="text-align:left">Servers: {{. | len}}</th>
    </tr>

	{{template "server-header"}}
	{{range . }}
		{{template "server-body" .}}
	{{end}}
</table>
`
