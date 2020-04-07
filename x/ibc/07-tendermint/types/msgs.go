package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Message types for the IBC client
const (
	TypeMsgCreateClient             string = "create_client"
	TypeMsgUpdateClient             string = "update_client"
	TypeMsgSubmitClientMisbehaviour string = "submit_client_misbehaviour"
)

var (
	_ clientexported.MsgCreateClient     = MsgCreateClient{}
	_ clientexported.MsgUpdateClient     = MsgUpdateClient{}
	_ evidenceexported.MsgSubmitEvidence = MsgSubmitClientMisbehaviour{}
)

// NewMsgCreateClient creates a new MsgCreateClient instance
func NewMsgCreateClient(
	id string, header Header,
	trustingPeriod, unbondingPeriod time.Duration, signer sdk.AccAddress,
) MsgCreateClient {
	return MsgCreateClient{
		ClientID:        id,
		Header:          header,
		TrustingPeriod:  trustingPeriod,
		UnbondingPeriod: unbondingPeriod,
		Signer:          signer,
	}
}

// GetClientID implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetClientID() string {
	return msg.ClientID
}

// Route implements sdk.Msg
func (msg MsgCreateClient) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgCreateClient) Type() string {
	return TypeMsgCreateClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgCreateClient) ValidateBasic() error {
	if msg.TrustingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidTrustingPeriod, "duration cannot be 0")
	}
	if msg.UnbondingPeriod == 0 {
		return sdkerrors.Wrap(ErrInvalidUnbondingPeriod, "duration cannot be 0")
	}
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	// ValidateBasic of provided header with self-attested chain-id
	if err := msg.Header.ValidateBasic(); err != nil {
		return sdkerrors.Wrapf(ErrInvalidHeader, "header failed validate basic with its own chain-id: %v", err)
	}
	return host.DefaultClientIdentifierValidator(msg.ClientID)
}

// GetSignBytes implements sdk.Msg
func (msg MsgCreateClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgCreateClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// GetClientType implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetClientType() string {
	return clientexported.ClientTypeTendermint
}

// GetConsensusState implements clientexported.MsgCreateClient
func (msg MsgCreateClient) GetConsensusState() clientexported.ConsensusState {
	// Construct initial consensus state from provided Header
	root := commitmenttypes.NewMerkleRoot(msg.Header.SignedHeader.GetHeader().GetAppHash())
	return &ConsensusState{
		Timestamp:    msg.Header.GetTime(),
		Root:         root,
		Height:       msg.Header.GetHeight(),
		ValidatorSet: msg.Header.GetValidatorSet(),
	}
}

// NewMsgUpdateClient creates a new MsgUpdateClient instance
func NewMsgUpdateClient(id string, header Header, signer sdk.AccAddress) MsgUpdateClient {
	return MsgUpdateClient{
		ClientID: id,
		Header:   header,
		Signer:   signer,
	}
}

// GetClientID implements clientexported.MsgUpdateClient
func (msg MsgUpdateClient) GetClientID() string {
	return msg.ClientID
}

// GetHeader implements clientexported.MsgUpdateClient
func (msg MsgUpdateClient) GetHeader() clientexported.Header {
	return &msg.Header
}

// Route implements sdk.Msg
func (msg MsgUpdateClient) Route() string {
	return ibctypes.RouterKey
}

// Type implements sdk.Msg
func (msg MsgUpdateClient) Type() string {
	return TypeMsgUpdateClient
}

// ValidateBasic implements sdk.Msg
func (msg MsgUpdateClient) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkerrors.ErrInvalidAddress
	}
	return host.DefaultClientIdentifierValidator(msg.ClientID)
}

// GetSignBytes implements sdk.Msg
func (msg MsgUpdateClient) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgUpdateClient) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// NewMsgSubmitClientMisbehaviour creates a new MsgSubmitClientMisbehaviour
// instance.
func NewMsgSubmitClientMisbehaviour(evidence evidenceexported.Evidence, signer sdk.AccAddress) MsgSubmitClientMisbehaviour {
	ev, ok := evidence.(*Evidence)
	if !ok {
		// TODO: panic or return error?
		return MsgSubmitClientMisbehaviour{}
	}
	return MsgSubmitClientMisbehaviour{
		Evidence:  ev,
		Submitter: signer,
	}
}

// Route returns the MsgSubmitClientMisbehaviour's route.
func (msg MsgSubmitClientMisbehaviour) Route() string { return ibctypes.RouterKey }

// Type returns the MsgSubmitClientMisbehaviour's type.
func (msg MsgSubmitClientMisbehaviour) Type() string { return TypeMsgSubmitClientMisbehaviour }

// ValidateBasic performs basic (non-state-dependant) validation on a MsgSubmitClientMisbehaviour.
func (msg MsgSubmitClientMisbehaviour) ValidateBasic() error {
	if msg.Evidence == nil {
		return sdkerrors.Wrap(evidencetypes.ErrInvalidEvidence, "missing evidence")
	}
	if msg.Submitter.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Submitter.String())
	}

	return msg.Evidence.ValidateBasic()
}

// GetSignBytes returns the raw bytes a signer is expected to sign when submitting
// a MsgSubmitClientMisbehaviour message.
func (msg MsgSubmitClientMisbehaviour) GetSignBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the single expected signer for a MsgSubmitClientMisbehaviour.
func (msg MsgSubmitClientMisbehaviour) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Submitter}
}

func (msg MsgSubmitClientMisbehaviour) GetEvidence() evidenceexported.Evidence {
	return msg.Evidence
}

func (msg MsgSubmitClientMisbehaviour) GetSubmitter() sdk.AccAddress {
	return msg.Submitter
}
