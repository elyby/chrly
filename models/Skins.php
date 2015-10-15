<?php

use Phalcon\Mvc\Collection;

/**
 * @property string $id
 */
class Skins extends Collection {

    public $_id;
    public $userId;
    public $nickname;
    public $skinId;
    public $url;
    public $is1_8;
    public $isSlim;
    public $hash;

    public function getId() {
        return $this->_id;
    }

    public function getSource() {
        return 'skins';
    }

    /**
     * @param string $nickname
     * @return bool|Skins
     */
    public static function findByNickname($nickname) {
        return static::findFirst([
            [
                'nickname' => mb_convert_case($nickname, MB_CASE_LOWER, ENCODING),
            ],
        ]);
    }

}
