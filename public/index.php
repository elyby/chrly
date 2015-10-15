<?php

use Phalcon\Mvc\Micro;

error_reporting(E_ALL);

try {
    /** @var \Phalcon\Config $config */
    $config = include __DIR__ . '/../config/config.php';
    /** @var \Phalcon\Loader $loader */
    include __DIR__ . '/../config/loader.php';
    /** @var Phalcon\DI\FactoryDefault $di */
    include __DIR__ . '/../config/services.php';

    $app = new Micro($di);
    include __DIR__ . '/../app.php';

    $app->handle();

} catch (Phalcon\Exception $e) {
    echo $e->getMessage();
} catch (PDOException $e) {
    echo $e->getMessage();
}
