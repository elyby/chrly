<?php

use Phalcon\Mvc\View;
use Phalcon\Mvc\Url as UrlResolver;
use Phalcon\DI\FactoryDefault;

$di = new FactoryDefault();

$di->set("view", function () {
    $view = new \Phalcon\Mvc\View();
    $view->disable();

    return $view;
});

/**
 * The URL component is used to generate all kind of urls in the application
 */
$di->set("url", function () use ($config) {
    $url = new UrlResolver();
    $url->setBaseUri($config->application->baseUri);

    return $url;
});

$di->set("mongo", function() use ($config) {
    if (!$config->mongo->username || !$config->mongo->password) {
        $mongo = new MongoClient(
            "mongodb://".
            $config->mongo->host.":".
            $config->mongo->port
        );
    } else {
        $mongo = new MongoClient(
            "mongodb://".
            $config->mongo->username.":".
            $config->mongo->password."@".
            $config->mongo->host.":".
            $config->mongo->port
        );
    }

    return $mongo->selectDb($config->mongo->dbname);
});

//Registering the collectionManager service
$di->setShared('collectionManager', function() {
    $modelsManager = new Phalcon\Mvc\Collection\Manager();
    return $modelsManager;
});