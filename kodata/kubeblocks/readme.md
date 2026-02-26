
#### mysql
secretName : mysql-a-mysql-account-root
内网： mysql-a-mysql
连接: mysql://<USERNAME>:<PASSWORD>@<HOST>:<PORT>/mysql
host IP.
        kubectl port-forward -n kb-db service/mysql-a-mysql 3306:3306

Endpoints:
COMPONENT   SERVICE-NAME    TYPE        PORT   INTERNAL                                EXTERNAL   
mysql       mysql-a-mysql   ClusterIP   3306   mysql-a-mysql.kb-db.svc.cluster.local              

Account Secrets:
COMPONENT   SECRET-NAME                  USERNAME   PASSWORD-KEY   
mysql       mysql-a-mysql-account-root   root       <password>     

========= connection example =========
========= .net connection example =========
# appsettings.json
{
  "ConnectionStrings": {
    "Default": "server=<HOST>;port=<PORT>;database=mysql;user=<USERNAME>;password=<PASSWORD>;SslMode=VerifyFull;"
  },
}

#### postgresql

secretName : pg-b-postgresql-account-postgres
内网： pg-b-postgresql-postgresql.kb-db.svc
连接: postgresql://<USERNAME>:<PASSWORD>@<HOST>:<PORT>/postgres
# you can use the following command to forward the service port to your local machine for testing the connection, using 127.0.0.1 as the host IP.
        kubectl port-forward -n kb-db service/pg-b-postgresql-postgresql 5432:5432

Endpoints:
COMPONENT    SERVICE-NAME                 TYPE        PORT        INTERNAL                                             EXTERNAL   
postgresql   pg-b-postgresql-postgresql   ClusterIP   5432,6432   pg-b-postgresql-postgresql.kb-db.svc.cluster.local              

Account Secrets:
COMPONENT    SECRET-NAME                        USERNAME   PASSWORD-KEY   
postgresql   pg-b-postgresql-account-postgres   postgres   <password>


### redis
secretName : redis-c-redis-account-default
内网： redis-c-redis-redis.kb-db.svc
# you can use the following command to forward the service port to your local machine for testing the connection, using 127.0.0.1 as the host IP.
        kubectl port-forward -n kb-db service/redis-c-redis-redis 6379:6379

Endpoints:
COMPONENT        SERVICE-NAME                            TYPE        PORT    INTERNAL                                                        EXTERNAL   
redis            redis-c-redis-redis                     ClusterIP   6379    redis-c-redis-redis.kb-db.svc.cluster.local                                
redis-sentinel   redis-c-redis-sentinel-redis-sentinel   ClusterIP   26379   redis-c-redis-sentinel-redis-sentinel.kb-db.svc.cluster.local              

Account Secrets:
COMPONENT        SECRET-NAME                              USERNAME   PASSWORD-KEY   
redis            redis-c-redis-account-default            default    <password>     
redis-sentinel   redis-c-redis-sentinel-account-default   default    <password> 



#### mongodb
secretName : mongo-d-mongodb-account-root
内网： mongo-d-mongodb.kb-db.svc
连接地址: mongodb://<USERNAME>:<PASSWORD>@<HOST>/admin
# you can use the following command to forward the service port to your local machine for testing the connection, using 127.0.0.1 as the host IP.
        kubectl port-forward -n kb-db service/mongo-d-mongodb 27017:27017

Endpoints:
COMPONENT   SERVICE-NAME                 TYPE        PORT    INTERNAL                                             EXTERNAL   
mongodb     mongo-d-mongodb              ClusterIP   27017   mongo-d-mongodb.kb-db.svc.cluster.local                         
mongodb     mongo-d-mongodb-mongodb      ClusterIP   27017   mongo-d-mongodb-mongodb.kb-db.svc.cluster.local                 
mongodb     mongo-d-mongodb-mongodb-ro   ClusterIP   27017   mongo-d-mongodb-mongodb-ro.kb-db.svc.cluster.local              

Account Secrets:
COMPONENT   SECRET-NAME                    USERNAME   PASSWORD-KEY   
mongodb     mongo-d-mongodb-account-root   root       <password>     

========= connection example =========
========= cli connection example =========
# mongodb client connection example
mongosh mongodb://<USERNAME>:<PASSWORD>@<HOST>/admin
