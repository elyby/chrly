<?php

$loader = new \Phalcon\Loader();

$loader->registerDirs(array(
    $config->application->modelsDir
));

$loader->register();
