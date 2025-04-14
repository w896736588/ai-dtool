package ai_service_tpl

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"strings"
)

func Service(cacheType, mainTemplateField, childTemplateField string) ([]_struct.Message, []_struct.Tool, error) {
	classList := make([]string, 0)
	classList = append(classList, ServiceClass())
	classList = append(classList, `}`)
	descList := []string{
		`你是一个php开发者，根据模板生成service，下面是示例，注意示例中的[]包起来的是提示,类名取值和注释应该基于历史会话生成的model类名`,
		`示例php controller:` + strings.Join(classList, "\n"),
		`当前缓存模式为：` + define.CacheTypeMap[cacheType],
		`注意最终输出结果要移除给你的这种[xxx]提示`,
	}
	if cacheType != define.NoCache {
		descList = append(descList, `注意：$main_template_field换为$`+mainTemplateField+`，代码中涉及到的要根据[]的提示进行处理`)
	}
	if cacheType == define.HashAdminCustomCache {
		descList = append(descList, `注意：$child_template_field为$`+childTemplateField+`，代码中涉及到的要根据[]的提示进行处理`)
	}
	return []_struct.Message{
		{
			Role:    define.RoleUser,
			Content: strings.Join(descList, `。`),
		},
	}, []_struct.Tool{}, nil
}

func ServiceClass() string {
	return `<?php

/**
 * [替换为表备注]
 * [注意TEMPLATE_REDIS_KEY 替换为model业务名]
 * Created by PhpStorm.
 * User: frog
 * Date: 2024/5/16 下午2:34
 */
class TemplateService extends BaseService
{
    //[model中的字段，凡是备注中有  2 1 0 之类的枚举字段都要生成const]
    const TYPE_TIMEOUT_CLOSE = 1; //超时自动结束接待
    const TYPE_VISITOR_NO_RESPONSE_CLOSE = 2; //访客无响应自动结束接待

    /**
     * 创建
     * User: frog
     * Date: 2024/5/16 下午3:38
     * @return array
     * @throws Exception
     */
    public function create($admin_user_id , $data)
    {
        $model = new TemplateModel();
		//[注意 根据分表模式 如果model中有setTableName 那么需要调用此方法并传递对应的参数]
        $id = $data['id'] ?? 0;
        $validate_ret = $this->validateForm($admin_user_id , $data , $id);
        if(is_error_code($validate_ret)){
            return $validate_ret;
        }
        if(empty($id)){ //创建
            //[以下字段按照model中的字段来,注意强转类型，根据字段类型来定]
            $model->create([
                'admin_user_id' => (int)$admin_user_id,
                'type' => (int)$data['type'],
                'staff_user_id' => (int)$data['staff_user_id'],
                'staff_group_id' => (int)$data['staff_group_id'],
                'staff_user_ids' => (string)$data['staff_user_ids'],
                'staff_group_ids' => (string)$data['staff_group_ids'],
                'create_time' => time(),
                'update_time' => time(),
            ]);
        }else{
            $old_config = $model->getOne(compact('admin_user_id' , 'id'));
            if(empty($old_config)){
                return code_err('不存在的规则');
            }
            //[以下字段按照model中的字段来,注意强转类型，根据字段类型来定]
            //[除了admin_user_id create_time id 其他的都需要更新]
            $model->updateOne([
                'type' => (int)$data['type'],
                'staff_user_id' => (int)$data['staff_user_id'],
                'staff_group_id' => (int)$data['staff_group_id'],
                'staff_user_ids' => (string)$data['staff_user_ids'],
                'staff_group_ids' => (string)$data['staff_group_ids'],
                'update_time' => time(),
            ] , compact('admin_user_id' , 'id'));
        }
        //[如果是hash自定义缓存 那么使用这下面两行]
        //[注意 这里的 main_template_field需要被换，如果它是admin_user_id，那么就不要这个参数了]
        $cache = $this->getHashCutomHandle($admin_user_id , $data['main_template_field']);
        $cache->delAll();
        //[如果存在string多条缓存 那么使用这下面两行]
        //[注意 这里的 $admin_user_id需要被换做main_template_field]
        $cache = $this->getStringsHandle();
        $cache->delete($admin_user_id);
        //[如果存在string单条缓存 那么使用这下面两行]
        //[注意 这里的 $admin_user_id需要被换做main_template_field]
        $cache = $this->getStringHandle();
        $cache->delete($admin_user_id);
        return code_ok();
    }

    /**
     * 表单验证
     * User: frog
     * Date: 2025/2/24 12:29
     * @return array
     */
    private function validateForm($admin_user_id , $data , $id){
        //验证id
        if(!empty($id)){
            $validate_id_ret = $this->validateId($admin_user_id , $id);
            if(is_error_code($validate_id_ret)){
                return $validate_id_ret;
            }
        }
        //[枚举字段都要进行验证,这里自行往下面加]
        if(isset($data['type'])){
            if(!in_array($data['type'] , [self::TYPE_TIMEOUT_CLOSE , self::TYPE_VISITOR_NO_RESPONSE_CLOSE])){
                return code_err('[这里换成这个字段的备注]验证失败');
            }
        }

        //[如果存在staff_user_id那么进行以下验证]
        if(!empty($data['staff_user_id'])){
            $num = (new Staff())->getStaffCountByIdList($admin_user_id , [$data['staff_user_id']]);
            if(empty($num)){
                return code_err('错误的客服');
            }
        }
        //[如果存在staff_user_ids那么进行以下验证]
        if(!empty($data['staff_user_ids'])){
            $num = (new Staff())->getStaffCountByIdList($admin_user_id , split_string_filter_unique($data['staff_user_id']));
            if(empty($num)){
                return code_err('错误的客服');
            }
        }
        //[如果存在staff_group_id那么进行以下验证]
        if(!empty($data['staff_group_id'])){
            $group_list = (new StaffGroupInfoModel())->getStaffGroupBySet([$data['staff_group_id']] , $admin_user_id) ?: [];
            if(empty($group_list)){
                return code_err('错误的客服组');
            }
        }
        //[如果存在staff_group_ids那么进行以下验证]
        if(!empty($data['staff_group_ids'])){
            $group_list = (new StaffGroupInfoModel())->getStaffGroupBySet(split_string_filter_unique($data['staff_group_ids']) , $admin_user_id) ?: [];
            if(empty($group_list)){
                return code_err('错误的客服组');
            }
        }
        //[如果存在wechatapp_id那么进行以下验证]
        if(!empty($data['wechatapp_id'])){
            $app_info = AppFactory::getOneByWechatId($data['wechatapp_id']);
            if(empty($app_info)){
                return code_err('错误的应用');
            }
            if($app_info['user_id'] != $admin_user_id){
                return code_err('错误的应用');
            }
            //[如果存在channel_id那么进行以下验证]
            if(!empty($data['channel_id'])){
                $channel_id_list = array_keys($app_info['channels']);
                if(!in_array($data['channel_id'] , $channel_id_list)){
                    return code_err('错误的渠道');
                }
            }
        }
        return code_ok();
    }

    /**
     * 列表展示返回
     * User: frog
     * Date: 2024/5/16 下午4:15
     * @throws Exception
     */
    public function getList($admin_user_id)
    {
        $model = new TemplateModel();
		//[注意 根据分表模式 如果model中有setTableName 那么需要调用此方法并传递对应的参数]
        //[注意这里的字段都要替换成model中的字段]
        $fields = 'id,admin_user_id';
        $list = $model->getAll(compact('admin_user_id') , $fields , 'id desc') ?: [];
        if(empty($list)){
            return [];
        }
        $list = $this->formatList($admin_user_id , $list);
        return code_ok('' , compact('list'));
    }

    /**
     * 格式化列表
     * User: frog
     * Date: 2025/2/24 12:25
     * @return array
     * @throws Exception
     */
    private function formatList($admin_user_id , $list)
    {
        //[如果存在字段staff_user_ids ，那么FormatServices::STAFF_NAME_LIST]
        //[如果存在字段staff_group_ids ，那么FormatServices::GROUP_NAME_LIST]
        //[如果存在字段staff_user_id ，那么FormatServices::STAFF_USER_NAME FormatServices::STAFF_NAME]
        //[如果存在字段create_time,那么FormatServices::CREATE_TIME]
        $format_service = new FormatServices($list , [FormatServices::GROUP_NAME_LIST , FormatServices::STAFF_NAME_LIST] , $admin_user_id);
        return $format_service->format()->getResult();
    }

    /**
     * 获取单个明细
     * User: frog
     * Date: 2024/5/16 下午5:05
     * @return array
     * @throws Exception
     */
    public function getDetail($admin_user_id , $id)
    {
        $validate_id_ret = $this->validateId($admin_user_id , $id);
        if(is_error_code($validate_id_ret)){
            return $validate_id_ret;
        }
        $config = $validate_id_ret['data']['config'];
        if(empty($config)){
            return code_err('不存在的配置');
        }
        $config = $this->formatList($admin_user_id , [$config])[0];
        return code_ok('' , $config);
    }

    /**
     * 删除单个规则
     * User: frog
     * Date: 2024/5/16 下午5:05
     * @return array
     * @throws Exception
     */
    public function delete($admin_user_id , $id)
    {
        $validate_id_ret = $this->validateId($admin_user_id , $id);
        if(is_error_code($validate_id_ret)){
            return $validate_id_ret;
        }
        $config = $validate_id_ret['data']['config'];
        //[如果是hash自定义缓存 那么使用这下面两行]
        //[注意 这里的 main_template_field需要被换，如果它是admin_user_id，那么就不要这个参数了]
        $cache = $this->getHashCutomHandle($admin_user_id , $config['main_template_field']);
        $cache->delAll();
        //[如果存在string多条缓存 那么使用这下面两行]
        //[注意 这里的 $admin_user_id需要被换做main_template_field]
        $cache = $this->getStringsHandle();
        $cache->delete($admin_user_id);
        //[如果存在string单条缓存 那么使用这下面两行]
        //[注意 这里的 $admin_user_id需要被换做main_template_field]
        $cache = $this->getStringHandle();
        $cache->delete($admin_user_id);
        return code_ok();
    }

    /**
     * 从缓存获取配置 [注意，如果是hash自定义缓存，那么创建一个以下的方法]
     * [注意 如果 我告诉你的main_template_field 是 admin_user_id 那么就移除掉$main_template_field，否则把$main_template_field换为我告诉你的值]
     * [注意 child_template_field换为我告诉你的字段名]
     * [注意 $admin_user_id是必有的]
     * User: frog
     * Date: 2024/5/16 下午4:04
     * @return array
     */
    public function getFromHashCustomCache($admin_user_id , $main_template_field , $child_template_field)
    {
        $cache = $this->getHashCutomHandle($admin_user_id , $main_template_field);
        if($cache->exists($cache->getOne($child_template_field))){
            $cache_data = $cache->getOne($child_template_field) ?: '';
            return json_decode($cache_data , true) ?: [];
        }else{
            $model = new TemplateModel();
			//[注意 根据分表模式 如果model中有setTableName 那么需要调用此方法并传递对应的参数]
            $config = $model->getOne(compact('admin_user_id' , 'main_template_field', 'child_template_field') , '*') ?: [];
            $cache->createPro([$child_template_field => json_encode($config)] , getCacheRand());
            return $config;
        }
    }

    /**
     * 从缓存获取列表
     * [注意，如果缓存类型是string多条缓存，那么创建一个以下的方法]
     * [注意 如果 我告诉你的main_template_field 是 admin_user_id 那么就移除掉$main_template_field，否则把$main_template_field换为我告诉你的值]
     * [注意 $admin_user_id是必有的]
     * User: frog
     * Date: 2024/5/16 下午4:04
     * @return array
     * @throws Exception
     */
    public function getFromStringsCache($admin_user_id , $main_template_field)
    {
        $cache = $this->getStringsHandle();
        return KeyCache::getCacheWithoutSerialize($cache , $main_template_field , function ()use($admin_user_id , $main_template_field){
            $model = new TemplateModel();
			//[注意 根据分表模式 如果model中有setTableName 那么需要调用此方法并传递对应的参数]
            return $model->getAll(compact('admin_user_id' , 'main_template_field') , '*');
        });
    }

    /**
     * 从缓存获取
     * [注意，如果是缓存类型为string单条缓存，那么创建一个以下的方法]
     * [注意 如果 我告诉你的main_template_field 是 admin_user_id 那么就移除掉$main_template_field，，否则把$main_template_field换为我告诉你的值]
     * [注意 如果 $main_template_field 就是 admin_user_id 那么就移除掉$main_template_field]
     * [注意 $admin_user_id是必有的]
     * User: frog
     * Date: 2024/5/16 下午4:04
     * @return array
     * @throws Exception
     */
    public function getFromStringCache($admin_user_id , $main_template_field)
    {
        $cache = $this->getStringsHandle();
        return KeyCache::getCacheWithoutSerialize($cache , $admin_user_id , function ()use($admin_user_id , $main_template_field){
            $model = new TemplateModel();
			//[注意 根据分表模式 如果model中有setTableName 那么需要调用此方法并传递对应的参数]
            return $model->getOne(compact('admin_user_id' , 'main_template_field'));
        });
    }

    /**
     * 验证ID
     * User: frog
     * Date: 2024/5/16 下午3:43
     * @return array
     */
    public function validateId($admin_user_id , $id)
    {
        if(empty($id)){
            return code_err('id不能为空');
        }
        $model = new TemplateModel();
		//[注意 根据分表模式 如果model中有setTableName 那么需要调用此方法并传递对应的参数]
        $config = $model->getOne(compact('admin_user_id' , 'id'));
        if(empty($config)){
            return code_err('错误的规则ID');
        }
        return code_ok('' , compact('config'));
    }

    /**
     * hash自定义缓存
     * [如果存在hash自定义缓存 那么生成这个方法]
     * [注意 $main_template_field需要替换为缓存主字段]
     * User: frog
     * Date: 2024/5/16 下午3:51
     * @return HashCache
     */
    public function getHashCutomHandle($admin_user_id , $main_template_field)
    {
        return new HashCache(getRedisKeyFromParam('TEMPLATE_REDIS_KEY' , $admin_user_id . '.' . $main_template_field));
    }

    /**
     * string多条缓存
     * [如果存在string多条缓存 那么生成这个方法]
     * User: frog
     * Date: 2024/5/16 下午3:51
     * @return KeyCache
     */
    public function getStringsHandle()
    {
        return new KeyCache(getRedisKeyFromParam('TEMPLATE_REDIS_KEY'));
    }

    /**
     * string单条缓存
     * [如果存在string单条缓存 那么生成这个方法]
     * User: frog
     * Date: 2024/5/16 下午3:51
     * @return KeyCache
     */
    public function getStringHandle()
    {
        return new KeyCache(getRedisKeyFromParam('TEMPLATE_REDIS_KEY'));
    }
}
`
}
