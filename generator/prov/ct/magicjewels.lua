local scripts = arg[0]:match("^(.*generator[/%\\])")
dofile(scripts.."lib/makereel.lua")

local symsetbase = {
	1, -- 1 wild (2, 3, 4 reels only)
	3, -- 2 crown    5000
	4, -- 3 ruby     1500
	5, -- 4 diamond  500
	6, -- 5 emerald  200
	6, -- 6 amber    50
	6, -- 7 sapphire 50
	6, -- 8 amethyst 40
}

local symsetfall = {
	1, -- 1 wild (2, 3, 4 reels only)
	8, -- 2 crown    5000
	7, -- 3 ruby     1500
	5, -- 4 diamond  500
	4, -- 5 emerald  200
	4, -- 6 amber    50
	4, -- 7 sapphire 50
	4, -- 8 amethyst 40
}

local neighbours = {
	--1, 2, 3, 4, 5, 6, 7, 8,
	{ 2, 0, 0, 0, 0, 0, 0, 0,}, -- 1 wild (2, 3, 4 reels only)
	{ 0, 2, 0, 0, 0, 0, 0, 0,}, -- 2 crown
	{ 0, 0, 2, 0, 0, 0, 0, 0,}, -- 3 ruby
	{ 0, 0, 0, 2, 0, 0, 0, 0,}, -- 4 diamond
	{ 0, 0, 0, 0, 2, 0, 0, 0,}, -- 5 emerald
	{ 0, 0, 0, 0, 0, 2, 0, 0,}, -- 6 amber
	{ 0, 0, 0, 0, 0, 0, 2, 0,}, -- 7 sapphire
	{ 0, 0, 0, 0, 0, 0, 0, 2,}, -- 8 amethyst
}

local function reelgen(n, isfall)
	local ss = tcopy(isfall and symsetfall or symsetbase)
	if n == 1 or n == 5 then
		ss[1] = 0
	end
	return makereel(ss, neighbours)
end

if autoscan then
	return reelgen
end

math.randomseed(os.time())
printreel(reelgen(1))
printreel(reelgen(2))
printreel(reelgen(3, true))
printreel(reelgen(4))
printreel(reelgen(5))
