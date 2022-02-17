# Jeen
Jeen is a package wrapper that is used as a web application base for the go language.

## Package Wrappers?

Yes, because jeen works by using other available packages.

## What's Different?

Jeen wrapped these packages to make usage shorter and code cleaner.
In addition, several new functions have been added which will certainly make the coding process faster.

Context cancellation has been included in every request so there is no need to check the timeout anymore. You only need to make sure the parts that don't use context.

Database and session has also been included so that it can be used easily in every http request. However, you can choose database and session drivers as needed.

## Credits
- Go Chi Router (https://github.com/go-chi/chi), we use chi as main router in this package with default middleware mentioned in documentation.
- SCS Session (https://github.com/alexedwards/scs), scs session is used but you can choose the driver for store.
- Scanny (https://github.com/georgysavva/scany), sqlscan dari scanny dikombinasikan dengan named parameter untuk memudahkan penulisan query.
sqlscan from scanny is combined with named parameters to make it easier to write queries.
- GoView (https://github.com/foolin/goview), we do not call goview explicitly but use the source code in the package and modify it for the needs of html and text templates.


## Author

This Package is developed by [Fuad Ar-Radhi](https://github.com/fuadarradhi).
