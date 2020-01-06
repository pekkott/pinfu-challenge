from wsgiref.simple_server import make_server
from calculate_hands import check_pinfu

import cgi, json
 
def app(environ, start_response):
  headers = [
    ('Content-type', 'application/json; charset=utf-8'),
    ('Access-Control-Allow-Origin', '*'),
  ]
  request_method = environ.get("REQUEST_METHOD")

  if request_method != "POST":
    status = '405 NG'
    start_response(status, headers)
    return []

  wsgi_input = environ["wsgi.input"]
  content_length = int(environ.get('CONTENT_LENGTH', 0))
  query = json.loads(wsgi_input.read(content_length))
  print(query)

  is_pinfu = check_pinfu(**query)

  status = '200 OK'
  start_response(status, headers)
 
  return is_pinfu
 
with make_server('', 8000, app) as httpd:
  print("Serving on port 8000...")
  httpd.serve_forever()
