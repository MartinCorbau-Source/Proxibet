# Proxibet
Pari entre pro

## Front (Angular)

Le projet Angular se trouve dans le dossier [proxifront](proxifront). Il faut se placer dans ce dossier avant de lancer les commandes (`npm start`, `ng serve`, `npm run lint`, etc.) :

```
cd proxifront
npm start
```


## Back (Golang)

```
docker compose up --build -d 
Get-Content .\proxiback\migrations\000001_create_users.up.sql | docker compose exec -T database psql -U proxibet -d proxibet
```