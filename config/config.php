<?php

return new \Phalcon\Config(array(
    'mongo' => array(
        'host'       => 'localhost',
        'port'       => 27017,
        'username'   => '',
        'password'   => '',
        'dbname'     => 'ely_skins',
    ),
    'application' => array(
        'modelsDir'      => __DIR__ . '/../models/',
        'baseUri'        => '/',
    )
));
