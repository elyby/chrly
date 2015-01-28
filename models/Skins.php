<?php

use Phalcon\Mvc\Collection;

/**
 * @method static Skins findFirst()
 *
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
        return "skins";
    }
} 