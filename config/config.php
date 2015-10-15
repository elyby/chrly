<?php

return new \Phalcon\Config([
    'mongo' => [
        'host'     => 'localhost',
        'port'     => 27017,
        'username' => '',
        'password' => '',
        'dbname'   => 'ely_skins',
    ],
    'application' => [
        'modelsDir' => __DIR__ . '/../models/',
        'baseUri'   => '/',
    ]
]);
