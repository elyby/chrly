<?php

define('ENCODING', 'UTF-8');

$app->get('/skins/{nickname}', function ($nickname) use ($app) {
    $systemVersion = $app->request->get('version', 'int');
    $minecraftVersion = $app->request->get('minecraft_version', 'string');

    // На всякий случай проверка на наличие .png для файла
    if (strrpos($nickname, '.png') != -1) {
        $nickname = explode('.', $nickname)[0];
    }

    // TODO: восстановить функцию деградации скинов

    $skin = Skins::findFirst(array(array(
        'nickname' => mb_convert_case($nickname, MB_CASE_LOWER, ENCODING)
    )));

    if (!$skin || $skin->skinId == 0) {
        return $app->response->redirect('http://skins.minecraft.net/MinecraftSkins/' . $nickname . '.png', true);
    }

    return $app->response->redirect($skin->url);
})->setName('skinSystem');

$app->get('/cloaks/{nickname}', function ($nickname) use ($app) {
    // На всякий случай проверка на наличие .png для файла
    if (strrpos($nickname, '.png') != -1) {
        $nickname = explode('.', $nickname)[0];
    }

    return $app->response->redirect('http://skins.minecraft.net/MinecraftCloaks/'.$nickname.'.png');
});

$app->get('/textures/{nickname}', function($nickname) use ($app) {
    $skin = Skins::findFirst(array(array(
        'nickname' => mb_convert_case($nickname, MB_CASE_LOWER, ENCODING)
    )));

    if ($skin && $skin->skinId != 0) {
        $url = $skin->url;
        $hash = $skin->hash;
    } else {
        $url = 'http://skins.minecraft.net/MinecraftSkins/'.$nickname.'.png';
        $hash = md5('non-ely-'.mktime(date('H'), 0, 0).'-'.$nickname);
    }

    $textures = array(
        'SKIN' => array(
            'url' => $url,
            'hash' => $hash,
            'metadata' => array(
                'model' => ($skin && $skin->isSlim) ? 'slim' : 'default'
            )
        ),
        'CAPE' => array(
            'url' => '',
            'hash' => ''
        )
    );

    return $app->response->setJsonContent($textures);
});

$app->post('/system/setSkin', function() use ($app) {
    $headers = getallheaders();
    if (!array_key_exists('X-Ely-key', $headers) || $headers['X-Ely-key'] != '43fd2ce61b3f5704dfd729c1f2d6ffdb') {
        return $app->response->setStatusCode(403, 'Forbidden')->setContent('Хорошая попытка, мерзкий хакер.');
    }

    $request = $app->request;
    $nickname = mb_convert_case($request->getPost('nickname', 'string'), MB_CASE_LOWER, ENCODING);

    $skin = Skins::findFirst(array(array(
        'nickname' => $nickname
    )));

    if (!$skin) {
        $skin = new Skins();
        $skin->nickname = $nickname;
    }

    $skin->userId =  (int) $request->getPost('userId', 'int');
    $skin->skinId = (int) $request->getPost('skinId', 'int');
    $skin->hash = $request->getPost('hash', 'string');
    $skin->is1_8 = (bool) $request->getPost('is1_8', 'int');
    $skin->isSlim = (bool) $request->getPost('isSlim', 'int');
    $skin->url = $request->getPost('url', 'string');

    if ($skin->save()) {
        echo 'OK';
    } else {
        echo 'ERROR';
    }
});

/**
 * Not found handler
 */
$app->notFound(function () use ($app) {
    $app->response
        ->setStatusCode(404, 'Not Found')
        ->setContent('Not Found<br /> <a href="http://ely.by">Система скинов Ely.by</a>.')
        ->send();
});
