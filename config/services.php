<?php
/**
 * @var \Phalcon\Config $config
 */

use Phalcon\Mvc\Collection\Manager;
use Phalcon\Mvc\View;
use Phalcon\Mvc\Url as UrlResolver;
use Phalcon\DI\FactoryDefault;

$di = new FactoryDefault();

$di->set('view', function () {
    $view = new View();
    $view->disable();

    return $view;
});

/**
 * The URL component is used to generate all kind of urls in the application
 */
$di->set('url', function () use ($config) {
    $url = new UrlResolver();
    $url->setBaseUri($config->application->baseUri);

    return $url;
});

$di->set('mongo', function() use ($config) {
    /** @var StdClass $mongoConfig */
    $mongoConfig = $config->mongo;
    $connectionString = 'mongodb://';
    if ($mongoConfig->username && $mongoConfig->password) {
        $connectionString .= "{$mongoConfig->username}:{$mongoConfig->password}@";
    }

    $connectionString .= $mongoConfig->host . ':' . $mongoConfig->port;
    $mongo = new MongoClient($connectionString);

    return $mongo->selectDb($mongoConfig->dbname);
});

$di->setShared('collectionManager', function() {
    return new Manager();
});
