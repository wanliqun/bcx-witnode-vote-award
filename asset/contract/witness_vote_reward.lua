-- local private function declaration
local func_datetime_from_string
local func_diff_days_between_dates
local func_calculate_reward_plan

-- constants
local asset_symbol='COCOS'

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

-- op_item: block_num/trx_id/voter/votee/amount/asset/timestamp
function add_vote_op_item(op_json)
    assert(chainhelper:is_owner(),'chainhelper:is_owner()')

    op_item=cjson.decode(op_json)
    -- read public data
    read_list={public_data={is_init=true,is_locked=true,is_cached=true,_vote_id=true,_vote_op_items=true}}
    chainhelper:read_chain()

    -- check validitity
    assert(public_data.is_init==true,'public_data.is_init==true')
    assert(public_data.is_locked==false,'public_data.is_locked==false')
    assert(op_item,'op_item~=nil')
    assert(op_item.votee==public_data._vote_id,'op_item.votee==public_data._vote_id')

    -- prepare voter op items
    local vote_op_items=public_data._vote_op_items
    local voter_items=vote_op_items[op_item.voter]
    if not(voter_items) 
    then
        vote_op_items[op_item.voter]={}
        voter_items=vote_op_items[op_item.voter]
    end

    -- check duplicates
    assert(voter_items[op_item.timestamp]==nil,'voter_items[op_item.timestamp]==nil')
    voter_items[op_item.timestamp]=op_item
    -- invalidate cache
    public_data.is_cached=false

    -- write public data
    write_list={public_data={is_cached=true,_vote_op_items=true}}
    chainhelper:write_chain()
end

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

-- calculate reward plan
func_calculate_reward_plan=function(context)
    local start_dt_str=context.calc_start_date
    local end_dt_str=context.calc_end_date
    local vote_op_items=context.vote_op_items

    local votes_dist={}
    local sum_total_votes=0

    for voter, op_items in pairs(vote_op_items) do
        -- sort timestamp first
        local tsset={}
        for k, v in pairs(op_items) do
            table.insert(tsset,k)
        end
        table.sort(tsset)

        local cur_dt,prev_dt,op_item,prev_op_item
        local total_votes = 0

        local num_items=#tsset
        table.insert(tsset,end_dt_str) -- also need to consider the ending range
        for i, itr_dt_str in pairs(tsset) do
            local cur_dt_str=itr_dt_str
            if (i<=num_items) then
                op_item=op_items[cur_dt_str]
            else
                op_item=nil
            end

            -- bound check
            if (itr_dt_str <= start_dt_str) then
                cur_dt_str = start_dt_str
            else 
                if (itr_dt_str > end_dt_str) then
                    cur_dt_str = end_dt_str
                end
            end

            cur_dt=func_datetime_from_string(cur_dt_str)
            if (prev_dt ~= nil and prev_op_item ~= nil) then 
                local y1,m1,d1,y2,m2,d2=prev_dt.year,prev_dt.month,prev_dt.day,cur_dt.year,cur_dt.month,cur_dt.day
                local count_days=func_diff_days_between_dates(y1,m1,d1,y2,m2,d2)
                
                -- do not count the current day on as the voting not finshed yet.
                count_days=count_days-1
                if (prev_dt.hour < 8) then -- count the previous datetime 
                    count_days=count_days+1
                end
                if (cur_dt.hour < 8) then
                    count_days=count_days-1
                end

                count_days=math.max(count_days,0)
                total_votes=total_votes+count_days*(prev_op_item.amount/100000)
            end

            prev_dt=cur_dt
            prev_op_item=op_item
        end

        votes_dist[voter]=total_votes
        sum_total_votes=sum_total_votes+total_votes
    end

    reward_plan={}
    if (sum_total_votes ~= 0) then
        for voter, num_votes in pairs(votes_dist) do
            reward_plan[voter]=num_votes/sum_total_votes
        end
    end

    return reward_plan
end

-- get date object from string formatted like '2019-09-12 05:23:30'
func_datetime_from_string=function(dt_str)
    local y, m, d, h, mt, s = string.match(dt_str, '(%d%d%d%d)-(%d%d)-(%d%d) (%d%d):(%d%d):(%d%d)$')
    return {year=y+0, month=m+0, day=d+0, hour=h+0, minute=mt+0, second=s+0}
end

-- days per month for normal or leap years
local _month_days={{31,31},{28,29},{31,31},{30,30},{31,31},{30,30},{31,31},{31,31},{30,30},{31,31},{30,30},{31,31}}
-- check if year is leap
local func_is_leap_year=function(year)
    return (year %4 == 0 and year % 100 ~= 0) or (year % 400 == 0)
end
-- calculate difference days between two date
func_diff_days_between_dates=function(y1,m1,d1,y2,m2,d2)
    local ans=0
    while (y1 < y2 or m1 < m2 or d1 < d2)
    do
        d1=d1+1

        local is_leap=func_is_leap_year(y1)
        local index=1
        if (is_leap) then 
            index=2 
        end

        if (d1==(_month_days[m1][index] + 1)) then
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