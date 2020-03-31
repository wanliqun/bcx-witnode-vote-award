# bcx-witnode-vote-award

bcx-witnode-vote-award（CocosBCX witness node vote award）是一个出块竞选节点投票奖励发放工具。

## 一、功能介绍

- 解析区块获取所有 CocosBCX 链上见证节点投票操作，并保存到数据库;
- 投票奖励通过智能合约进行发放，转账结算算法公开透明。

## 二、使用说明

### 1. 导入 MySQL 数据库

将 asset/dbschema/bcx_wit_votes.sql 文件导入 MySQL 数据库

### 2. 修改配置文件

根据配置注释修改 config.yaml 文件：

```yaml
cocosbcx:
  node: # cocosbcx node settings
    host: 127.0.0.1 # node host
    port: 41346 # node rpc port
    use_ssl: false # use ssl or not

  wallet: # sdk wallet settings
    wallet_path: ./wallet.dat # wallet load/save path
    wallet_pwd: CocosBCX is awesome! # wallet password
    wif_prk: 5JHdMwsWkEXsMozFrQAQKnKwo44CaV77H45S9PsH7QVbFQngJfw # wif private key
    bcx_account: kokko # use bcx account
    
mysql:
  host: 127.0.0.1 # mysql host
  port: 33060 # mysql port
  username: root # mysql username
  password: cocosbcx # mysql password 
  database: bcx_wit_votes # mysql database
```

### 3. 部署智能合约

将 asset/contract/witness_vote_reward.lua 智能合约文件部署到 CocosBCX 公链上，具体部署方法可以参考 [CocosBCX 官方文档](http://https://cn-dev.cocosbcx.io/docs/57-%E5%90%88%E7%BA%A6%E5%BC%80%E5%8F%91%E5%92%8C%E9%83%A8%E7%BD%B2%E7%9A%84%E5%A4%9A%E7%A7%8D%E6%96%B9%E5%BC%8F "官方文档")。

### 4. 初始化智能合约

调用智能合约中的 init 方法初始化智能合约，其中调用参数说明如下：
- vote_id - 配置需要进行奖励发放的见证节点投票 id（如 1:22, 1:0 等，具体id可通过查看见证节点详情来获取）
- start_date - 需要进行统计奖励的起始日期，如 "2019-03-25 00:00:00"
- end_date - 需要进行统计奖励的截至日期，如 "2019-03-31 23:00:00"

```lua
function init(vote_id,start_date,end_date)
    assert(chainhelper:is_owner(),'chainhelper:is_owner()')
    assert(vote_id,'vote_id~=nil')
    assert(start_date,'start_date~=nil')
    assert(end_date,'end_date~=nil')
    assert(start_date<end_date,'start_date<end_date')

    -- read public data
    read_list={public_data={is_init=true}}
    chainhelper:read_chain()
    assert(public_data.is_init==nil,'public_data.is_init==nil')

    -- init public data 
    public_data.is_init=true
    public_data.is_locked=false
    public_data.is_cached=false

    public_data._vote_id=vote_id
    public_data._start_date=start_date
    public_data._end_date=end_date

    public_data._vote_op_items={}
    public_data._reward_plans={}
    public_data._bonus_distrib={}

    -- write public data
    chainhelper:write_chain()
end
```

### 5. 抓取区块、解析投票操作

通过运行 golang fetch 命令从配置的节点中抓取区块并进行投票操作解析：

`go run main.go fetch --config config.yaml -s 100000 -e 1200000`

>  其中的 -s 参数指定抓取的起始块高度，-e 参数指定抓取的截至块高度；如果不指定的话，默认从创世块开始抓取，抓取到最新块高度时会在前台等待几秒后继续进行抓取

抓取区块后解析的操作会实时写入 MySQL 数据库，入库时已做去重处理可重复抓取。

### 6. 投票操作记录上传智能合约

通过运行 golang submmit 命令将抓取区块中解析的操作提交到 CocosBCX 智能合约中：

`go run main.go submmit --config config.yaml -v 1:22 -c contract.vote_reward -s "2020-02-15 00:00:00" -e"2020-03-25 00:00:00"`

>  其中的 -s 和 -e 是可选参数，可指定指定时间范围进行的投票操作进行提交，分别表示起始和截至时间，无配置时默认提交所有时间范围内的见证节点投票操作；-v 和 -c 是必选参数，-v 指定被投票见证节点的 vote-id, 必须和初始化智能合约时指定的 vote-id 一致，否则提交时会出现 assert 失败，-c 指定智能合约的名称

submmit 命令将从 MySQL 数据库中读出符合滤选条件的投票操作记录，通过调用智能合约中的add_vote_op_item 方法将投票操作记录保存到智能合约公有数据 public_data 下的 _vote_op_items 对象中。

### 7. 预览奖励发放计划

在发放奖励之前，可通过调用智能合约的 peek_reward_plan 方法预览奖励发放计划，peek_reward_plan 将计算所有用户的投票数量以及投票周期次数计算投票占比，奖励发放将根据这个比例进行发放。

```lua
function peek_reward_plan()
    -- read public data
    read_list={public_data={is_init=true,is_cached=true}}
    chainhelper:read_chain()

    -- check validility
    assert(public_data.is_init==true,'public_data.is_init==true')

    -- calculate reward plan or get it in cache directly
    if (not public_data.is_cached) then
        read_list={public_data={_start_date=true,_end_date=true,_vote_op_items=true}}
        chainhelper:read_chain()

        -- do the calculation
        local context={
            calc_start_date=public_data._start_date,
            calc_end_date=public_data._end_date,
            vote_op_items=public_data._vote_op_items
        }
        local reward_plans=func_calculate_reward_plan(context)
        public_data._reward_plans=reward_plans
        public_data.is_cached=true

        -- write reward plans to chain
        write_list={public_data={is_cached=true,_reward_plans=true}}
        chainhelper:write_chain()
    end
end
```
调用完智能合约 peek_reward_plan 方法后，具体的分配方案将保存在智能合约公有数据 public_data 下的 _reward_plans 对象中，下次在没有添加新的投票操作前提下重复调用此方法将直接使用缓存的数据对象。

### 8. 发放奖励

可通过智能合约的 settle_pay_reward 方法来进行最终的奖励发放，它将计算奖励分配比例（如有必要参考peek_reward_plan 方法），并将指定数量的 token 按照相应比例进行转账到每个投票账号中。

```lua
function settle_pay_reward(reward_bonus)
    assert(chainhelper:is_owner(),'chainhelper:is_owner()')

    reward_bonus=reward_bonus+0
    assert(reward_bonus>0,'reward_bonus>0')

     -- read public data
     read_list={public_data={is_init=true,is_locked=true,is_cached=true,_reward_plans=true}}
     chainhelper:read_chain()
 
     -- check validitity
     assert(public_data.is_init==true,'public_data.is_init==true')
     assert(public_data.is_locked==false,'public_data.is_locked==false')

     -- calculate reward plan or get it in cache directly
    local reward_plan=public_data._reward_plans
    if (not public_data.is_cached) then
        read_list={public_data={_start_date=true,_end_date=true,_vote_op_items=true}}
        chainhelper:read_chain()
        
        local context={
            calc_start_date=public_data._start_date,
            calc_end_date=public_data._end_date,
            vote_op_items=public_data._vote_op_items
        }
        reward_plan=func_calculate_reward_plan(context)
        public_data._reward_plans=reward_plan
        public_data.is_cached=true
    end

    -- transfer & distribute bonus 
    local bonus_distrib={}
    for voter, share in pairs(reward_plan) do
        local amount=math.floor(reward_bonus*share)
        if (amount > 0) then
            chainhelper:transfer_from_caller(voter,amount,asset_symbol,true)
            bonus_distrib[voter]=amount
        end
    end
    public_data._bonus_distrib=bonus_distrib
    public_data.is_locked=true

    -- write back to chain
    write_list={public_data={is_locked=true,is_cached=true,_reward_plans=true,_bonus_distrib=true}}
    chainhelper:write_chain()

    -- output result
    local json_bonus=cjson.encode(bonus_distrib)
    chainhelper:log('settle_pay_reward'..',date:'..date('%Y-%m-%dT%H:%M:%S',chainhelper:time())..',json_bonus:'..json_bonus)
end
```
settle_pay_reward 方法调用成功后，具体的分配方案将保存在智能合约公有数据 public_data下的 bonus_distrib 对象下，并同时写入交易日志中，并最终锁定智能合约。

## 三、奖励分配算法

为了减少复杂度，bcx-witnode-vote-award 采用以下简化模型来进行投票奖励分配比例计算。在投票奖励时间范围内的所有的投票不分其他权重设置，只考虑投票的数量和投票时间周期的次数。

以一个简单实例来说明下，假设投票奖励时间范围设置为 "2020-01-15 00:00:00" 到  "2020-03-15 00:00:00"，见证节点投票 id 为 1.22。在此期间附近用户 1.2.100 对见证节点有如下投票记录：

- "2020-01-14 07:02:00" 投票为 5000 COCOS
- "2020-02-18 15:02:00" 投票为 18000 COCOS

用户 1.2.105对见证节点有如下投票记录：

- “2020-02-10 10:30:00” 投票为 8000 COCOS
- “2020-03-05 15:02:00” 投票为 12000 COCOS

**1.2.100 的总投票数计算方式为：**
"2020-01-14 07:02:00"（实际以"2020-01-15 00:00:00"开始计算） 到 "2020-02-18 15:02:00" 时间段投票周期为 33 次（按每天8点为链维护周期时间来计算），此段时间的投票总数为 5000 x 33 = 165000； "2020-03-05 15:02:00" 到 "2020-03-15 00:00:00" 时间段投票周期为 14 次, 此段时间的投票总数为 18000 x 14 = 252000;

**1.2.105 的总投票数计算方式为：**
“2020-02-10 10:30:00”到 “2020-03-05 15:02:00” 时间段投票周期为 22 次，此段时间的投票总数为 8000 x 22 = 176000; “2020-03-05 15:02:00”到 "2020-03-15 00:00:00" 时间段投票周期为 8 次, 此段时间的投票总数为 12000 x 8 = 96000;

1.2.100 总的投票次数为：165000+252000=417000；
1.2.105 总的投票次数为：176000+96000=272000;
1.2.100 的分配比例为： 417000/(417000+272000) = 0.6052249637155298；
1.2.105 的分配比例为： 272000/(417000+272000) = 0.39477503628447025
