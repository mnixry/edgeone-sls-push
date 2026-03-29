# edgeone-sls-push

A lightweight HTTP bridge that receives [Tencent EdgeOne](https://www.tencentcloud.com/products/edgeone) real-time log pushes and forwards them to [Alibaba Cloud Log Service (SLS)](https://www.alibabacloud.com/product/log-service).

## How it works

1. EdgeOne pushes logs via HTTP POST to a configurable endpoint (default `/edgeone-logs`).
2. The service verifies the request using EdgeOne query-string authentication (`auth_key` + `access_key`).
3. The JSON (or JSON Lines) body is parsed, normalized, and enqueued to SLS via the producer SDK.
4. A health check endpoint (`GET /healthz`) verifies SLS connectivity.

## Quick start

### Run with Go

```bash
go build -o app .
./app \
  --edgeone-secret-id="$EDGEONE_SECRET_ID" \
  --edgeone-secret-key="$EDGEONE_SECRET_KEY" \
  --sls-endpoint="cn-hangzhou.log.aliyuncs.com" \
  --sls-access-key-id="$SLS_ACCESS_KEY_ID" \
  --sls-access-key-secret="$SLS_ACCESS_KEY_SECRET" \
  --sls-project="my-project" \
  --sls-logstore="my-logstore"
```

### Run with Docker

```bash
docker run -p 8080:8080 <IMAGE_NAME> \
  --edgeone-secret-id="$EDGEONE_SECRET_ID" \
  --edgeone-secret-key="$EDGEONE_SECRET_KEY" \
  --sls-endpoint="cn-hangzhou.log.aliyuncs.com" \
  --sls-access-key-id="$SLS_ACCESS_KEY_ID" \
  --sls-access-key-secret="$SLS_ACCESS_KEY_SECRET" \
  --sls-project="my-project" \
  --sls-logstore="my-logstore"
```

All flags can also be set via environment variables (uppercased, with `_` separators and a section prefix):

```bash
export EDGEONE_SECRET_ID="..."
export EDGEONE_SECRET_KEY="..."
export SLS_ENDPOINT="cn-hangzhou.log.aliyuncs.com"
export SLS_ACCESS_KEY_ID="..."
export SLS_ACCESS_KEY_SECRET="..."
export SLS_PROJECT="my-project"
export SLS_LOGSTORE="my-logstore"
./app
```

## Configuration

### Required

| Environment variable    | Flag                      | Description                                        |
| ----------------------- | ------------------------- | -------------------------------------------------- |
| `EDGEONE_SECRET_ID`     | `--edgeone-secret-id`     | EdgeOne SecretId for signature verification        |
| `EDGEONE_SECRET_KEY`    | `--edgeone-secret-key`    | EdgeOne SecretKey for signature verification       |
| `SLS_ENDPOINT`          | `--sls-endpoint`          | SLS endpoint (e.g. `cn-hangzhou.log.aliyuncs.com`) |
| `SLS_ACCESS_KEY_ID`     | `--sls-access-key-id`     | Alibaba Cloud AccessKey ID                         |
| `SLS_ACCESS_KEY_SECRET` | `--sls-access-key-secret` | Alibaba Cloud AccessKey Secret                     |
| `SLS_PROJECT`           | `--sls-project`           | SLS project name                                   |
| `SLS_LOGSTORE`          | `--sls-logstore`          | SLS logstore name                                  |

### Optional

| Environment variable  | Flag                    | Default         | Description                                   |
| --------------------- | ----------------------- | --------------- | --------------------------------------------- |
| `HTTP_ADDR`           | `--http-addr`           | `:8080`         | Listen address                                |
| `HTTP_PATH`           | `--http-path`           | `/edgeone-logs` | URL path for log pushes                       |
| `HTTP_READ_TIMEOUT`   | `--http-read-timeout`   | `30s`           | HTTP read timeout                             |
| `HTTP_WRITE_TIMEOUT`  | `--http-write-timeout`  | `30s`           | HTTP write timeout                            |
| `HTTP_MAX_BODY_BYTES` | `--http-max-body-bytes` | `10485760`      | Max request body size (bytes)                 |
| `EDGEONE_MAX_SKEW`    | `--edgeone-max-skew`    | `300s`          | Max allowed clock skew for auth timestamp     |
| `SLS_TOPIC`           | `--sls-topic`           |                 | SLS log topic                                 |
| `SLS_SOURCE`          | `--sls-source`          |                 | SLS log source                                |
| `SLS_LINGER_MS`       | `--sls-linger-ms`       | `2000`          | Batch linger time (ms)                        |
| `SLS_MAX_BATCH_SIZE`  | `--sls-max-batch-size`  | `524288`        | Max batch size (bytes)                        |
| `SLS_MAX_BATCH_COUNT` | `--sls-max-batch-count` | `4096`          | Max logs per batch                            |
| `SLS_RETRIES`         | `--sls-retries`         | `10`            | Max retry attempts                            |
| `LOG_LEVEL`           | `--log-level`           | `info`          | Log level                                     |
| `LOG_FORMAT`          | `--log-format`          | `json`          | Log format (`json` or `console`)              |
| `LOG_OUTPUT`          | `--log-output`          | `stdout`        | Log output (`stdout`, `stderr`, or file path) |

## License

MIT, see [LICENSE](LICENSE) for details.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
    OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
    SOFTWARE.
