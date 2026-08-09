jaisheesh@LAPTOP-2U6VV3L0:/mnt/c/Users/Balagowni Jaisheesh/Desktop/Projects/MCP-Go$ docker compose down -v
[+] down 2/2
 ✔ Container mcp-go-mysql-1 Removed                                                    1.7s
 ✔ Network mcp-go_default   Removed                                                    0.4s
jaisheesh@LAPTOP-2U6VV3L0:/mnt/c/Users/Balagowni Jaisheesh/Desktop/Projects/MCP-Go$ go test ./integration -tags=integration -v
=== RUN   TestIntegration_IntrospectAndMergeAgainstExamples
[mysql] 2026/08/09 10:59:51 packets.go:58 unexpected EOF
[mysql] 2026/08/09 10:59:52 packets.go:58 unexpected EOF
[mysql] 2026/08/09 10:59:53 packets.go:58 unexpected EOF
[mysql] 2026/08/09 10:59:54 packets.go:58 unexpected EOF
[mysql] 2026/08/09 10:59:55 packets.go:58 unexpected EOF
[mysql] 2026/08/09 10:59:56 packets.go:58 unexpected EOF
[mysql] 2026/08/09 10:59:57 packets.go:58 unexpected EOF
[mysql] 2026/08/09 10:59:58 packets.go:58 unexpected EOF
[mysql] 2026/08/09 10:59:59 packets.go:58 unexpected EOF
[mysql] 2026/08/09 11:00:00 packets.go:58 unexpected EOF
time=2026-08-09T11:00:01.295+05:30 level=INFO msg="context file loaded" tables=16
time=2026-08-09T11:00:01.309+05:30 level=INFO msg="schema introspected" time=2026-08-09T05:30:01Z tables=16 columns=168
time=2026-08-09T11:00:01.309+05:30 level=INFO msg="schema introspected" tables=16
time=2026-08-09T11:00:01.309+05:30 level=INFO msg="schema merged" tables=16
--- PASS: TestIntegration_IntrospectAndMergeAgainstExamples (15.31s)
PASS
ok      github.com/Jaisheesh-2006/schema-context-mcp/integration        15.320s
jaisheesh@LAPTOP-2U6VV3L0:/mnt/c/Users/Balagowni Jaisheesh/Desktop/Projects/MCP-Go$ 