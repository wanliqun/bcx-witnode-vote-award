-- local private function declaration
local _date_from_string
local _diff_days
local _calculate_reward_plan

-- constants
local asset_symbol='COCOS'

local function init(vote_id)
    assert(chainhelper:is_owner(),'Must be the owner to be priviledged for init action')

    -- read public data
    read_list={public_data={is_init=true}}
    chainhelper:read_chain()
    assert(public_data.is_init==nil,'Already initialized')

    -- init public data 
    public_data.is_init=true
    public_data.is_locked=false
    public_data.is_cached=false

    public_data._vote_id=vote_id
    public_data._vote_op_items={}
    public_data._reward_plans={}
    public_data._bonus_distrib={}

    -- write public data
    chainhelper:write_chain()
end

-- op_item: block_num/trx_id/voter/votee/amount/asset/timestamp
local function add_vote_op_item(op_item)
    assert(chainhelper:is_owner(),'Must be the owner to be priviledged for add action')

    -- read public data
    read_list={public_data={is_init=true,is_locked=true,is_cached=true,_vote_id=true,_vote_op_items=true}}
    chainhelper:read_chain()

    -- check validitity
    assert(public_data.is_init==true,'This contract has not been initized yet')
    assert(public_data.is_locked==false,'This contract has already been locked')
    assert(op_item.votee==public_data._vote_id, 'Invalid vote operation (wrong vote id)')

    -- prepare voter op items
    local vote_op_items=public_data._vote_op_items
    local voter_items=vote_op_items[op_item.voter]
    if not(voter_items) 
    then
        vote_op_items[op_item.voter]={}
        voter_items=vote_op_items[op_item.voter]
    end

    -- check duplicates
    assert(voter_items[op_item.timestamp]==nil,'Voter operation item already exists')

    voter_items[op_item.timestamp]=op_item
    public_data.is_cached=false

    -- write public data
    write_list={public_data={is_cached=true,_vote_op_items=true}}
    chainhelper:write_chain()
end

local function peek_reward_plan()
    -- read public data
    read_list={public_data={is_init=true,is_cached=true,_vote_op_items=true,_reward_plans=true}}
    chainhelper:read_chain()

    -- check validility
    assert(public_data.is_init==true,'This contract has not been initized yet')

    -- calculate reward plan or get in cache 
    local reward_plan=public_data._reward_plans
    local vote_op_items=public_data._vote_op_items
    if (not public_data.is_cached) then
        reward_plan=_calculate_reward_plan(vote_op_items)
        public_data.is_cached=true

        -- write reward plans to chain
        write_list={public_data={is_cached=true,_reward_plans=true}}
        chainhelper:write_chain()
    end

    -- output result
    local json_plan=cjson.encode(reward_plan)
    chainhelper:log('func:peek_reward_plan'..',date:'..date('%Y-%m-%dT%H:%M:%S',chainhelper:time())..',json_plan:'..json_plan)
end

function settle_pay_reward(reward_bonus)
    assert(chainhelper:is_owner(),'Must be the owner to be priviledged for settle action')
    assert(reward_bonus>0,'Reward bonus amount must be greater than 0')

     -- read public data
     read_list={public_data={is_init=true,is_locked=true,is_cached=true,_vote_op_items=true,_reward_plans=true}}
     chainhelper:read_chain()
 
     -- check validitity
     assert(public_data.is_init==true,'This contract has not been initized yet')
     assert(public_data.is_locked==false,'This contract has already been locked')

     -- calculate reward plan or get in cache 
    local reward_plan=public_data._reward_plans
    local vote_op_items=public_data._vote_op_items
    if (not public_data.is_cached) then
        reward_plan=_calculate_reward_plan(vote_op_items)
        public_data.is_cached=true
    end

    -- transfer & distribute bonus 
    local bonus_distrib = public_data._bonus_distrib
    for voter, share in pairs(reward_plan) do
        local amount=reward_bonus*share
        if (amount > 0) then
            chainhelper:transfer_from_caller(voter,amount,asset_symbol,true)
            bonus_distrib[voter]=amount
        end
    end
    public_data.is_locked=true

    -- write reward plans to chain
    write_list={public_data={is_cached=true,_reward_plans=true,_bonus_distrib=true}}
    chainhelper:write_chain()

    -- output result
    local json_bonus=cjson.encode(bonus_distrib)
    chainhelper:log('func:settle_pay_reward'..',date:'..date('%Y-%m-%dT%H:%M:%S',chainhelper:time())..',json_bonus:'..json_bonus)
end

_calculate_reward_plan=function(vote_op_items)
    -- calculate total votes for each voter
    local votes_dist={}
    local sum_total_votes=0
    for voter, op_items in pairs(vote_op_items) do 
        table.sort(op_items)

        local prev_timestamp,prev_op_itme
        local total_votes = 0
        for timestamp, op_item in pairs(op_items) do
            local cur_timestamp
            if ( previous_timestamp) then 
                cur_timestamp=_date_from_string(timestamp)
                local count_days=_diff_days(prev_timestamp.year,prev_timestamp.month,prev_timestamp.day,cur_timestamp.year,cur_timestamp.month,cur_timestamp.day)
                -- not count the current timestamp day on
                count_days=count_days-1
                if (prev_timestamp.hour < 8) then
                    count_days=count_days+1
                end
                if (cur_timestamp.hour < 8) then 
                    count_days=count_days-1
                end

                count_days=math.max(count_days,0)
                total_votes=total_votes+count_days*(prev_op_itme.amount/100000)
            end
            prev_op_itme=op_item
            prev_timestamp=cur_timestamp
        end

        votes_dist[voter]=total_votes
        sum_total_votes=sum_total_votes+total_votes
    end

    -- calculate reward distribution plan
    local reward_plan={}
    for voter, num_votes in pairs(votes_dist) do
        reward_plan[voter]=num_votes/sum_total_votes
    end

    return reward_plan
end

-- get date object from string formatted like '2019-09-12 05:23:30'
_date_from_string=function(str)
    local y, m, d, h, mt, s = string.match(str, '(%d%d%d%d)-(%d%d)-(%d%d) (%d%d):(%d%d):(%d%d)$')
    return {year=y+0, month=m+0, day=d+0, hour=h+0, minute=mt+0, second=s+0}
end

-- days per month for average or leap years
local _month_days={{31,31},{28,29},{31,31},{30,30},{31,31},{30,30},{31,31},{31,31},{30,30},{31,31},{30,30},{31,31}}
-- check if year is leap
local is_leap_year=function(year)
    return (year %4 == 0 and year % 100 ~= 0) or (year % 400 == 0)
end
-- calculate difference days between two date
_diff_days = function(y1,m1,d1,y2,m2,d2)
    local ans=0
    while (y1 < y2 or m1 < m2 or d1 < d2)
    do
        d1=d1+1

        local is_leap=is_leap_year(y1)
        local index=1
        if (is_leap) then 
            index=2 
        end

        if (d1==(month_days[m1][index] + 1)) then
            m1=m1+1 -- 日期变为下个月1号
            d1=1
        end

        if (m1 == 13) then -- 月份满12个月
            y1=y1+1
            m1=1    --日期变为下一年的1月
        end

        ans=ans+1
    end

    return ans
end