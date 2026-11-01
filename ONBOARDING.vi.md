# Làm quen với codebase

Tài liệu này dành cho người mới vào dự án. Đọc xong bạn chạy được service trên máy mình, đi được
một request từ `curl` tới PostgreSQL và biết sửa ở đâu khi có yêu cầu mới. Thuật ngữ kỹ thuật giữ
nguyên tiếng Anh để khớp với tên trong code. Quy tắc đầy đủ nằm ở
[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), lý do của từng quyết định ở [docs/adr](docs/adr).

## 1. Bức tranh lớn trong một phút

Đây là một **modular monolith**: một binary, một database, một gRPC server, một HTTP server đứng
trước làm REST. Bên trong là nhiều **module** độc lập, mỗi module là một **hexagon** (ports and
adapters):

- `core/` là phần nghiệp vụ, chỉ dùng standard library. Linter chặn mọi import khác.
- `port/in` là những gì module **cho phép** người khác gọi. `port/out` là những gì module **cần**
  từ bên ngoài. Cả hai đều là interface.
- `adapter/in` biến request gRPC thành lời gọi vào `port/in`. `adapter/out` implement `port/out`
  bằng PostgreSQL.
- `internal/monolith` là chủ nhà: giữ server, database pool, và gọi `Register` của từng module.

Hiện có một module mẫu là `versioning`: trả về version của app, Go, PostgreSQL và schema.

## 2. Cài đặt và chạy lần đầu

Cần có: Go 1.26 trở lên, [buf](https://buf.build/docs/installation),
[golangci-lint](https://golangci-lint.run) v2, và một PostgreSQL. sqlc và goose chạy qua `go run`
với version ghim sẵn trong `Makefile`, không cần cài.

```sh
export DATABASE_URL='postgresql://user:password@localhost:5432/app?sslmode=disable'

make migrate          # tạo schema bằng goose CLI
make run              # gRPC :9090, HTTP :8080
```

Ở terminal khác:

```sh
curl -s localhost:8080/v1/versioning/version
curl -si localhost:8080/readyz | head -1
```

Nếu thấy JSON có `appVersion`, `databaseVersion`, `schemaVersion` là xong. Các biến môi trường
khác và giá trị mặc định xem ở [internal/config/.env.example](internal/config/.env.example).

## 3. Đi một request từ đầu tới cuối

Đọc theo đúng thứ tự này, mỗi file một lần, bạn sẽ hiểu toàn bộ đường chạy:

| Bước | File | Chuyện gì xảy ra |
|---|---|---|
| 1 | `cmd/app/main.go` | Đọc config, cấu hình `slog`, bắt SIGINT/SIGTERM, gọi `monolith.New` với danh sách module, rồi `Run`. |
| 2 | `internal/monolith/monolith.go` | `New`: mở pool PostgreSQL, tạo gRPC server với interceptor, tạo gateway mux, gọi `Register` của từng module, dựng HTTP server. `Run`: listen, chờ signal, shutdown có thứ tự. |
| 3 | `internal/monolith/interceptor.go` | `recoverUnary` đổi panic thành `codes.Internal`; `logUnary` log method, code, duration của mọi RPC. |
| 4 | `internal/versioning/module.go` | Wiring của module, đọc từ ngoài vào trong: driven adapter, use case, driving adapter, rồi đăng ký lên gRPC server và gateway. |
| 5 | `api/versioning/v1/versioning.proto` | Hợp đồng API. Dòng `option (google.api.http)` là thứ tạo ra route REST `GET /v1/versioning/version`. |
| 6 | `internal/versioning/adapter/in/grpc/versioning_handler.go` | Nhận `GetVersionRequest`, gọi `port/in.VersionService`, map `domain.Version` sang protobuf. Lỗi được log ở đây và client chỉ nhận thông báo chung. |
| 7 | `internal/versioning/core/port/in/version_service.go` | Interface `VersionService`. Handler chỉ biết interface này. |
| 8 | `internal/versioning/core/usecase/version_service.go` | Implement `VersionService`: gọi `port/out.DatabaseInfoReader` hai lần, lấy `runtime.Version()`, ráp thành `domain.Version`. |
| 9 | `internal/versioning/core/domain/version.go` | `Version` và value object `AppVersion`, tự validate semver khi tạo. |
| 10 | `internal/versioning/core/port/out/database_info.go` | Interface `DatabaseInfoReader`. Use case chỉ biết interface này. |
| 11 | `internal/versioning/adapter/out/postgres/database_info.go` | Implement `DatabaseInfoReader` bằng `db.Queries`. |
| 12 | `db/query/versioning.sql` và `internal/db/versioning.sql.go` | SQL viết tay và Go do sqlc sinh ra từ SQL đó. |

Request REST đi thêm một chặng ở đầu: HTTP server nhận, grpc-gateway dịch sang gRPC rồi gọi
qua loopback vào chính gRPC server của process. Nhờ vậy REST và gRPC đi cùng một interceptor.

## 4. Thư mục nào chứa gì

```
cmd/app/            entry point duy nhất
api/                file .proto và cấu hình buf; là nguồn sự thật của API
gen/                code buf sinh ra; không sửa tay
db/migrations/      goose migration, một lịch sử cho một database
db/schema/          DDL của bảng do tool tự tạo (bảng version của goose), chỉ để sqlc đọc
db/query/           SQL cho sqlc, mỗi module một file đặt theo tên module
internal/db/        code sqlc sinh ra; không sửa tay
internal/config/    đọc biến môi trường
internal/logging/   cấu hình slog
internal/monolith/  chủ nhà của các module
internal/<module>/  một bounded context, bên trong là core/ và adapter/
docs/               ARCHITECTURE.md và ADR
```

Quy tắc nhớ nhanh: thứ gì chung cho cả database thì nằm ở `db/`, thứ gì của riêng module thì nằm
trong `internal/<module>/`, thứ gì có header `Code generated` thì không sửa tay.

## 5. Từ điển thuật ngữ

| Thuật ngữ | Nghĩa trong dự án này |
|---|---|
| module | Một bounded context, là một thư mục `internal/<name>` có `module.go`. |
| core | Phần nghiệp vụ của module: `domain`, `port`, `usecase`. Chỉ import standard library. |
| domain | Entity và value object có hành vi, ví dụ `AppVersion` tự validate. |
| port/in, driving port | Interface module cung cấp cho bên ngoài gọi. |
| port/out, driven port | Interface module cần bên ngoài cung cấp, ví dụ đọc database. |
| usecase | Application service implement `port/in` bằng cách gọi `domain` và `port/out`. |
| adapter/in, driving adapter | Code dịch transport (gRPC) thành lời gọi `port/in`. |
| adapter/out, driven adapter | Code implement `port/out` bằng hạ tầng thật (PostgreSQL). |
| monolith | Package `internal/monolith`, giữ server và pool, gọi `Register` của module. |
| Deps | Struct chứa mọi thứ module được phép đăng ký lên: gRPC server, gateway mux, loopback conn, pool, logger. |
| grpc-gateway | Thư viện sinh REST handler từ annotation trong `.proto`, gọi gRPC qua loopback. |
| buf | Tool lint và generate cho protobuf; thay cho `protoc`. |
| sqlc | Tool đọc SQL viết tay trong `db/query/` và sinh Go có kiểu trong `internal/db/`. |
| goose | Tool migration. Chạy bằng CLI qua `make migrate`, app không tự migrate. |
| depguard | Rule của golangci-lint chặn `core/` import ngoài standard library. |
| function-field mock | Struct có một field `func` cho mỗi method của port, dùng trong unit test. Field chưa gán thì panic. |

## 6. Việc thường gặp và sửa ở đâu

**Thêm một RPC vào module có sẵn**
1. Thêm `rpc` và message vào `api/<module>/v1/<module>.proto`, kèm `google.api.http` nếu cần REST.
2. `make generate`.
3. Thêm method vào `port/in`, implement trong `usecase`, thêm field vào `inmock`.
4. Thêm method vào handler ở `adapter/in/grpc`.
5. Viết test cho usecase bằng `outmock`, cho handler bằng `inmock`.

**Thêm một query SQL**
1. Viết query vào `db/query/<module>.sql` theo cú pháp sqlc (`-- name: Foo :one`).
2. `make generate`, kiểm tra `internal/db/<module>.sql.go`.
3. Thêm method vào `port/out`, implement trong `adapter/out/postgres`, thêm field vào `outmock`.

**Thêm một migration**
```sh
make migrate-create NAME=add_orders_table
make migrate
```
Migration lên database **trước** khi code mới được deploy, nên phải tương thích với version app
đang chạy: thêm cột, thêm bảng thoải mái; xóa hay đổi tên chỉ làm sau khi code đã ngừng dùng một
release.

**Thêm một module mới**: xem mục "Adding a module" trong [README.md](README.md). Tóm tắt: copy
`internal/versioning`, thêm core của module vào allow list của `depguard` trong `.golangci.yml`,
tạo proto và query, viết `module.go`, thêm một dòng vào `cmd/app/main.go`.

## 7. Test

```sh
make test                      # unit test luôn chạy; test PostgreSQL chỉ chạy khi có DATABASE_URL
make lint                      # golangci-lint, gồm cả rule depguard
make generate && git diff      # CI fail nếu diff khác rỗng
```

- Unit test ở `core` và `adapter/in` dùng mock trong `inmock/` và `outmock/`. Trong test, gán hàm
  cho field cần dùng, không gán thì mock panic để bạn biết mình quên.
- Test ở `adapter/out/postgres` cần database thật đã chạy `make migrate`; không có `DATABASE_URL`
  thì test tự skip.

## 8. Lỗi hay gặp ngày đầu

| Triệu chứng | Nguyên nhân thường gặp |
|---|---|
| `config: DATABASE_URL is required` | Chưa `export DATABASE_URL`. |
| `make lint` báo depguard ở file trong `core/` | Bạn import thư viện ngoài vào core. Chuyển phần đó ra adapter, hoặc đưa qua `port/out`. |
| CI job `generate` đỏ | Sửa proto hoặc SQL nhưng chưa `make generate`, hoặc sửa tay file generated. |
| `SchemaVersion = 0` trong test | Database chưa `make migrate`. |
| Mock panic `...Fn is not set` | Test gọi tới method chưa gán hàm. Gán field tương ứng trong test. |
| `curl` REST trả 404 | Route REST sinh từ `google.api.http` trong proto; kiểm tra annotation rồi `make generate`. |

## 9. Checklist ngày đầu

- [ ] Chạy được `make migrate && make run` và `curl` thấy JSON.
- [ ] Đọc 12 file ở mục 3 theo đúng thứ tự.
- [ ] Chạy `make test` và `make lint` xanh trên máy mình.
- [ ] Thử thêm một field vào `GetVersionResponse` theo mục 6, chạy generate, test, rồi bỏ đi.
- [ ] Đọc `docs/ARCHITECTURE.md` một lượt; ADR đọc khi gặp câu hỏi "tại sao".
