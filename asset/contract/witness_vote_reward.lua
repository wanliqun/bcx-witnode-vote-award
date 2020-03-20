function init()
    assert(chainhelper:is_owner(), '!chainhelper:is_owner()')

    read_list={public_data={is_init=true}}
    chainhelper:read_chain()
    assert(public_data.is_init==nil,'public_data.is_init!=nil')

    public_data._vote_op_items = {}
    public_data._alloc_rewards = {}

    public_data.is_init=true
    chainhelper:write_chain()
end

function add_vote_op_item(block_num, trx_id, voter, votee, amount, asset, timestamp)
end

function settle()
end