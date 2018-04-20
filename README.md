# Webservice Client as Service Provider

CASP(client as service provider) will implements a server and a client, server does not own data and service, the client will support data for server.

that means, you will make a data service via server host, but data is located on client host, you do not want move data to server host, maybe your server host has no enough space, you can make a program connect to server, server forwarding request to this client, client read data from local and send to server host, then server finished the data service.

# Proto
- CASP server is a hollow server, no data restored.
- when a data request to CASP server, CASP server will forwarding to CASP client.
- CASP client maybe a web server, serve data service to local, an serve to CASP server.
- when CASP client start, CASP client connect to CASP server, tell server which data can be service, then wait for message from CASP server, message might be a http request, and then send to CASP server the http response. this will service a reverted request->response procedure.

# Supported Protocal or not
- HTTP/1.1
- HTTPS
- not HTTP/2
- not WebSocket
- not Proxy



# Server side
start http/websocket server and listen;
API of register a CASP service for CASP client;
API of request a service forwarding to CASP client;


# Client side
Connect to CASP server via websocket and register service to CASP server;
start a local http server as service;

# Optional plan
- Base on WebSocket
- Base on Tcp protocol
- Base on gRPC / both side stream
- Base on p2p protocal
