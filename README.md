# eBrick extensions

**eBrick Extensions** is a collection of modular components and utilities designed to extend the functionality of the eBrick framework. These extensions provide additional capabilities, integrations, and tools to accelerate development and deployment.

---

## Key Features

- **Plug-and-Play Design**: Easily integrate extensions into your existing eBrick-based projects.
- **Scalable Modules**: Components for caching, logging, database integration, messaging, and more.
- **Configurable**: Customizable to meet specific application requirements.
- **Open Source**: Built to foster collaboration and innovation within the developer community.

---

## Available Extensions

Below is a list of extensions currently available in this repository:

- **Cache**: Support for Redis, Memcached, and other caching mechanisms.
- **Database**: Integration with popular databases such as PostgreSQL and MySQL.
- **Logger**: Preconfigured logging solutions using Zap or other logging libraries.
- **Event**: Support for In-memory, Kafka, RabbitMQ event bus.
- **Configuration Management**: Flexible configuration loading using Viper and environment variables.

---
## Importing and Using eBrick Extensions
The eBrick Extensions library provides modular components that can be easily integrated into your project. Follow these steps to use specific extensions in your project.
### 1. Import the Extensions

To use a specific extension, simply import it in your Go file. For example, to use the cache, database, event, and logger extensions:

```go
import (
	_ "github.com/ebrickdev/extensions/v1/cache/gocache"
	_ "github.com/ebrickdev/extensions/v1/database/postgresql"
	_ "github.com/ebrickdev/extensions/v1/event/nats"
	_ "github.com/ebrickdev/extensions/v1/logger/logrus"
)
```
•	The _ (blank identifier) ensures that the extensions’ init() functions are executed, even if their exported functionality is not directly used in your code.

•	These extensions automatically register themselves within the eBrick framework for seamless integration.
