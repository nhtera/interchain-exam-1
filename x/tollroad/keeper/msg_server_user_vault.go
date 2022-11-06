package keeper

import (
	"context"

	"github.com/b9lab/toll-road/x/tollroad/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateUserVault(goCtx context.Context, msg *types.MsgCreateUserVault) (*types.MsgCreateUserVaultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value already exists
	_, isFound := k.GetUserVault(
		ctx,
		msg.Creator,
		msg.RoadOperatorIndex,
		msg.Token,
	)
	if isFound {
		return nil, sdkerrors.Wrap(types.ErrIndexSet, "index already set")
	}

	var userVault = types.UserVault{
		Owner:             msg.Creator,
		RoadOperatorIndex: msg.RoadOperatorIndex,
		Token:             msg.Token,
		Balance:           msg.Balance,
	}

	creatorAddr, _ := sdk.AccAddressFromBech32(msg.Creator)
	if msg.Balance == 0 {
		return nil, sdkerrors.Wrap(types.ErrZeroTokens, "zero tokens")
	}
	coin := sdk.NewCoin(msg.Token, sdk.NewInt(int64(msg.Balance)))
	errBank := k.bank.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, sdk.NewCoins(coin))
	if errBank != nil {
		return nil, errBank
	}

	k.SetUserVault(
		ctx,
		userVault,
	)
	return &types.MsgCreateUserVaultResponse{}, nil
}

func (k msgServer) UpdateUserVault(goCtx context.Context, msg *types.MsgUpdateUserVault) (*types.MsgUpdateUserVaultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	valFound, isFound := k.GetUserVault(
		ctx,
		msg.Creator,
		msg.RoadOperatorIndex,
		msg.Token,
	)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	// Checks if the the msg creator is the same as the current owner
	if msg.Creator != valFound.Owner {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	// If the balance field in the message is 0 then it returns an error, because conceptually, this should be a deletion.
	if msg.Balance == 0 {
		return nil, sdkerrors.Wrap(types.ErrZeroTokens, "zero tokens")
	}

	creatorAddr, _ := sdk.AccAddressFromBech32(msg.Creator)
	coinBalance := int64(msg.Balance)
	coinBalance2 := int64(valFound.Balance)

	// If the balance field in the message is higher than the current vault balance, the difference is transferred from the user to the module. And should return an error if it is not possible.
	if msg.Balance > valFound.Balance {
		coin2 := sdk.NewCoin(msg.Token, sdk.NewInt(coinBalance-coinBalance2))
		err := k.bank.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, sdk.NewCoins(coin2))
		if err != nil {
			return nil, err
		}
	} else {
		// If the balance field in the message is lower than the current vault balance, the difference is transferred from the module to the user. And should panic if it is not possible.
		coin := sdk.NewCoin(msg.Token, sdk.NewInt(coinBalance2-coinBalance))
		err := k.bank.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, sdk.NewCoins(coin))
		if err != nil {
			panic(err.Error())
		}
	}

	var userVault = types.UserVault{
		Owner:             msg.Creator,
		RoadOperatorIndex: msg.RoadOperatorIndex,
		Token:             msg.Token,
		Balance:           msg.Balance,
	}

	k.SetUserVault(ctx, userVault)

	return &types.MsgUpdateUserVaultResponse{}, nil
}

func (k msgServer) DeleteUserVault(goCtx context.Context, msg *types.MsgDeleteUserVault) (*types.MsgDeleteUserVaultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the value exists
	valFound, isFound := k.GetUserVault(
		ctx,
		msg.Creator,
		msg.RoadOperatorIndex,
		msg.Token,
	)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrKeyNotFound, "index not set")
	}

	// Checks if the the msg creator is the same as the current owner
	if msg.Creator != valFound.Owner {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	creatorAddr, _ := sdk.AccAddressFromBech32(msg.Creator)
	coinBalance := int64(valFound.Balance)
	coin3 := sdk.NewCoin(msg.Token, sdk.NewInt(coinBalance))
	err := k.bank.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, sdk.NewCoins(coin3))
	if err != nil {
		panic(err.Error())
	}

	k.RemoveUserVault(
		ctx,
		msg.Creator,
		msg.RoadOperatorIndex,
		msg.Token,
	)

	return &types.MsgDeleteUserVaultResponse{}, nil
}
