
```bash
docker build -t ccr.ccs.tencentyun.com/eqxiu/waf-to-es .
docker push ccr.ccs.tencentyun.com/eqxiu/waf-to-es

kubectl -n eqxiu-ops rollout restart deploy/waf-to-es-service
```