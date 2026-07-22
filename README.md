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
```

Les migrations SQL presentes dans `proxiback/migrations` sont appliquees automatiquement au demarrage de l'API.