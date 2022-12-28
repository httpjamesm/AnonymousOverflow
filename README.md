# AnonymousOverflow

AnonymousOverflow allows you to view StackOverflow threads without the cluttered interface and exposing your IP address, browsing habits and other browser fingerprint data to StackOverflow.

This project is super lightweight by design. The UI is simple and the frontend is served as an SSR HTML requiring no JavaScript.

## Screenshots

![Home](https://cdn.horizon.pics/7HbXKZ3JBJbBDMtt51DVJgo5THFMYX.png)

![Question](https://cdn.horizon.pics/lqEuJMzYpgeB8au4QASsZ0S0BQHknJ.png)

![Answer](https://cdn.horizon.pics/0RwQtrxH3COTC9sfo0FsQ56I3Opl0m.png)

## Clearnet Instances

| Instance URL                                         | Region                  | Notes                                                                                            |
| ---------------------------------------------------- | ----------------------- | ------------------------------------------------------------------------------------------------ |
| [code.whatever.social](https://code.whatever.social) | United States & Germany | Operated by [Whatever Social](https://whatever.social) and [http.james](https://httpjames.space) |
| [overflow.777.tf](https://overflow.777.tf/)          | The Netherlands         | Operated by [Jae](https://777.tf)                                                                |
| [ao.vern.cc](https://ao.vern.cc)                     | United States           | Operated by [vern.cc](https://vern.cc)                                                           |

## Other Instances

| Instance URL                                                                                                                                  | Region        | Notes                                  |
| --------------------------------------------------------------------------------------------------------------------------------------------- | ------------- | -------------------------------------- |
| [ao.vernccvbvyi5qhfzyqengccj7lkove6bjot2xhh5kajhwvidqafczrad.onion](http://ao.vernccvbvyi5qhfzyqengccj7lkove6bjot2xhh5kajhwvidqafczrad.onion) | United States | Operated by [vern.cc](https://vern.cc) |
| [vernmzgraj6aaoafmehupvtkkynpaa67rxcdj2kinwiy6konn6rq.b32.i2p](http://vernmzgraj6aaoafmehupvtkkynpaa67rxcdj2kinwiy6konn6rq.b32.i2p)           | United States | Operated by [vern.cc](https://vern.cc) |

## How it works

AnonymousOverflow uses the existing question endpoint that StackOverflow uses. Simply replace the domain name in the URL with the domain name of the AnonymousOverflow instance you're using and you'll be able to view the question anonymously.

Example:

```
https://stackoverflow.com/questions/43743250/using-libsodium-xchacha20-poly1305-for-large-files
```

becomes

```
${instanceURL}/questions/43743250/using-libsodium-xchacha20-poly1305-for-large-files
```

## How to deploy

AnonymousOverflow uses Docker for deployment.

Install Docker for your platform and copy the `docker-compose.example.yml` file to `docker-compose.yml`.

Then, tweak settings if needed.

Run `docker compose up -d` to build and start the container.

## Attribution

-   Icons provided by [heroicons](https://heroicons.com) under the [MIT License](https://choosealicense.com/licenses/mit/)
