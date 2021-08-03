from http.server import HTTPServer, SimpleHTTPRequestHandler


class MyHandler(SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.extensions_map[".wasm"] = "application/wasm"


httpd = HTTPServer(("", 8181), MyHandler)
httpd.serve_forever()
