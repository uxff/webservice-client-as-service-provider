# Webservice Client as Service Provider

CASP(client as service provider) will implements a server and a client, server does not own data and service, the client will support data for server.

that means, you will make a data service via server host, but data is located on client host, you do not want move data on server, maybe your server has no enough space, you can make a program connect to server, server routes request to this client, client read data from local and send to server host, then server finished the data service.

# Proto
- CASP server is a hollow server, no data restored.
- when a data request to CASP server, CASP server will forwarding to CASP client.
- CASP client maybe a web server, serve data service to local, an serve to CASP server.
- when CASP client start, CASP client connect to CASP server, tell server which data can be service, then wait for message from CASP server, message might be a http request, and then send to CASP server the http response. this will service a reverted request->response procedure.


