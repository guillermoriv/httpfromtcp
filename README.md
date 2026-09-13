# httpfromtcp

An implementation of HTTP built directly on top of raw TCP sockets in Go.

The goal of this project is to build an HTTP/1.1 server from scratch without relying on high-level libraries like `net/http`. By working directly with raw TCP connections, the focus is on understanding how the protocol operates under the hood: reading byte streams, parsing request lines and headers, handling message bodies, and formatting compliant HTTP responses.

Implemented according to:

- [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110.html) (HTTP Semantics)
- [RFC 9112](https://www.rfc-editor.org/rfc/rfc9112.html) (HTTP/1.1)
