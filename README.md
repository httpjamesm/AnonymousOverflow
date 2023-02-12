# AnonymousOverflow

AnonymousOverflow allows you to view StackOverflow threads without the cluttered interface and exposing your IP address, browsing habits and other browser fingerprint data to StackOverflow.

This project is super lightweight by design. The UI is simple and the frontend is served as an SSR HTML requiring no JavaScript.

## Screenshots

![Home](https://files.horizon.pics/e2b9275c-1409-4978-801b-de981a8d3ae9?a=1&mime1=image&mime2=png)

![Question](https://files.horizon.pics/0f6b0036-87f0-4acd-9a0f-936b5c397a73?a=1&mime1=image&mime2=png)

![Answer](https://files.horizon.pics/861ec510-644b-43f2-9439-0a2cae841422?a=1&mime1=image&mime2=png)

## Clearnet Instances

| Instance URL                                                                    | Region                  | Notes                                                                                            |
| ------------------------------------------------------------------------------- | ----------------------- | ------------------------------------------------------------------------------------------------ |
| [code.whatever.social](https://code.whatever.social)                            | United States & Germany | Operated by [Whatever Social](https://whatever.social) and [http.james](https://httpjames.space) |
| [overflow.777.tf](https://overflow.777.tf/)                                     | The Netherlands         | Operated by [Jae](https://777.tf)                                                                |
| [ao.vern.cc](https://ao.vern.cc)                                                | United States           | Operated by [vern.cc](https://vern.cc)                                                           |
| [overflow.smnz.de](https://overflow.smnz.de)                                    | Germany                 | Operated by [smnz.de](https://smnz.de)                                                           |
| [anonymousoverflow.esmailelbob.xyz](https://anonymousoverflow.esmailelbob.xyz/) | Canada                  | Operated by [Esmail EL BoB](https://esmailelbob.xyz)                                             |
| [overflow.lunar.icu](https://overflow.lunar.icu)                                | Germany                 | Operated by [lunar.icu](https://lunar.icu/)                                                      |
| [ao.foss.wtf](https://ao.foss.wtf)                                              | Germany                 | Operated by [foss.wtf](https://foss.wtf)                                                         |
| [overflow.hostux.net](https://overflow.hostux.net/)                             | France                  | Operated by [Hostux](https://hostux.net/)                                                        |

## Other Instances

| Instance URL                                                                                                                                                                | Region        | Notes                                                |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------- | ---------------------------------------------------- |
| [ao.vernccvbvyi5qhfzyqengccj7lkove6bjot2xhh5kajhwvidqafczrad.onion](http://ao.vernccvbvyi5qhfzyqengccj7lkove6bjot2xhh5kajhwvidqafczrad.onion)                               | United States | Operated by [vern.cc](https://vern.cc)               |
| [vernmzgraj6aaoafmehupvtkkynpaa67rxcdj2kinwiy6konn6rq.b32.i2p](http://vernmzgraj6aaoafmehupvtkkynpaa67rxcdj2kinwiy6konn6rq.b32.i2p)                                         | United States | Operated by [vern.cc](https://vern.cc)               |
| [anonymousoverflow.esmail5pdn24shtvieloeedh7ehz3nrwcdivnfhfcedl7gf4kwddhkqd.onion](http://anonymousoverflow.esmail5pdn24shtvieloeedh7ehz3nrwcdivnfhfcedl7gf4kwddhkqd.onion) | Canada        | Operated by [Esmail EL BoB](https://esmailelbob.xyz) |

## Why use AnonymousOverflow over StackOverflow?

-   StackOverflow collects a lot of information

While it's understandable that StackOverflow collects a lot of technical data for product development, debugging and to serve the best experience to its users, not everyone wants their

> internet protocol (IP) address, [...] browser type and version, time zone setting and location, browser plug-in types and versions, operating system, and platform [...] data

to be collected and stored.

-   StackOverflow shares your information with third-parties

StackOverflow does not sell your information, but it does share it with third-parties, including conglomerates.

> We also partner with other third parties, such as Google Ads and Microsoft Bing, to serve advertising content and manage advertising campaigns. When we use Google Ads or Microsoft Bing Customer Match for advertising campaigns, your personal data will be protected using hashed codes.
> Google users can control the ads that they see on Google services, including Customer Match ads, in their Google Ads Settings.

Their main website also [contains trackers from Alphabet](https://themarkup.org/blacklight?url=stackoverflow.com).

-   Reduced clutter

StackOverflow has a cluttered UI that might distract you from the content you're trying to find. AnonymousOverflow simplifies the interface to make it easier to read and navigate.

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

### Bookmark Conversion Tool

You can easily convert StackOverflow URLs to AnonymousOverflow ones by adding the following code as a bookmark in your web browser:

```js
javascript: (function () {
    window.location = window.location
        .toString()
        .replace(/stackoverflow\.com/, "code.whatever.social");
})();
```

Replace `code.whatever.social` with the domain name of the instance you're using if needed.

You can run this bookmarklet on any StackOverflow page to view it anonymously.

Thanks to [Manav from ente.io](https://ente.io/about) for the handy tool.

## How to deploy

Read the [wiki page](https://github.com/httpjamesm/AnonymousOverflow/wiki/Deployment).

## Attribution

-   Icons provided by [heroicons](https://heroicons.com) under the [MIT License](https://choosealicense.com/licenses/mit/)
-   [Gin](https://github.com/gin-gonic/gin) under the [MIT License](https://github.com/gin-gonic/gin/blob/master/LICENSE)
-   [goquery](https://github.com/PuerkitoBio/goquery) under the [BSD 3-Clause License](https://github.com/PuerkitoBio/goquery/blob/master/LICENSE)
-   [resty](https://github.com/go-resty/resty) under the [MIT License](https://github.com/go-resty/resty/blob/master/LICENSE)
-   [Chroma](https://github.com/alecthomas/chroma) under the [MIT License](https://github.com/alecthomas/chroma/blob/master/COPYING)
