curl -X POST -H "Content-Type: application/json" -d '{
"destination": "192.168.254.0/24",
"gateway": "192.168.10.1",
"interface": "lo0",
"metric": "1"
}' http://localhost:8080/routes
