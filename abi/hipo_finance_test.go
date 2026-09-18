package abi

import (
	"encoding/hex"
	"testing"

	"github.com/tonkeeper/tongo/boc"
)

// Every message below is a real one taken from mainnet, one per Hipo op declared in
// hipo_finance.xml that this file covers. They are here because the op-codes are only half
// the schema: a field declared at the wrong width still decodes, it just reports the wrong
// number. TestHipoFinanceFieldWidths below pins the two that are easy to get wrong.
func TestHipoFinanceMessages(t *testing.T) {
	cases := []struct {
		name string
		ext  bool
		body string
		want MsgOpName
	}{
		{name: "participate_in_election", ext: true, body: "B5EE9C72010101010012000020574A297B000000006A9E2B966A9E4F08", want: HipoFinanceParticipateInElectionExtInMsgOp},
		{name: "vset_changed", ext: true, body: "B5EE9C720101010100120000202F0B5B3B000000006AA24F296AA14F08", want: HipoFinanceVsetChangedExtInMsgOp},
		{name: "finish_participation", ext: true, body: "B5EE9C7201010101001200002023274435000000006A95CF676A944F08", want: HipoFinanceFinishParticipationExtInMsgOp},
		{name: "decide_loan_requests", ext: false, body: "B5EE9C720101010100120000206A31D344000000006A9E2B966A9E4F08", want: HipoFinanceDecideLoanRequestsMsgOp},
		{name: "process_loan_requests", ext: false, body: "B5EE9C72010101010012000020071D07CC000000006A9E2B966A9E4F08", want: HipoFinanceProcessLoanRequestsMsgOp},
		{name: "recover_stakes", ext: false, body: "B5EE9C720101010100120000204F173D3E000000006A95CF676A944F08", want: HipoFinanceRecoverStakesMsgOp},
		{name: "proxy_new_stake", ext: false, body: "B5EE9C7201010301009C000118089CD4D0000000006A9E2B960101904539B3CE286825E1027B3558C95D016AA774DD24C5E6238EA57F967C0937595D6A9E4F0800030000A86F0B341C62D351E454B8ABBD6A6A041FC9A901F795F5AD307E6A2EF5AF029F020080C7F9BA5F103E7ABF9F9A328C7390739F50DFFD83F6530561945C489CFAD46438AE1D1D68B38021C686AA2BBD814004D32F8E71F76B3AD4BA1E320AB874E70F00", want: HipoFinanceProxyNewStakeMsgOp},
		{name: "proxy_recover_stake", ext: false, body: "B5EE9C7201010101000E000018407CB243000000006A95CF67", want: HipoFinanceProxyRecoverStakeMsgOp},
		{name: "request_rejected", ext: false, body: "B5EE9C7201010101000E000018CD0F2116000000006AA92B87", want: HipoFinanceRequestRejectedMsgOp},
		{name: "take_profit", ext: false, body: "B5EE9C7201010101000E0000188B556813000000006A95CF67", want: HipoFinanceTakeProfitMsgOp},
		{name: "take_borrower_fee", ext: false, body: "B5EE9C7201010101000E0000185E2D81F4000000006A9FCF60", want: HipoFinanceTakeBorrowerFeeMsgOp},
		{name: "request_loan", ext: false, body: "B5EE9C720101030100AF00013E36335DA9000000006AA5512B6AA64F087016BFFF2242C0052D7FFFD14007070101901781DB04A92CB9444B3043E9C8DB1BB6FC23B7DA38B9E7902A8ECCDEA2632F796AA64F08000300000F130A89779CC0D6E9CBA7571982050AF2AD2C59206879E5762DC21235D3BA66020080BFF1F56BB51083DB7B22CF827282C6D759FEBE75B5D011DEC36EF29933D7AC763757B03237DA9F88F85982223F4F91C9801C33631536342937184DD93D46FE06", want: HipoFinanceRequestLoanMsgOp},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw, err := hex.DecodeString(c.body)
			if err != nil {
				t.Fatal(err)
			}
			cells, err := boc.DeserializeBoc(raw)
			if err != nil {
				t.Fatal(err)
			}
			var op *MsgOpName
			if c.ext {
				_, op, _, err = ExtInMessageDecoder(cells[0], nil)
			} else {
				_, op, _, err = InternalMessageDecoder(cells[0], nil)
			}
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if op == nil {
				t.Fatalf("not decoded, so the op is still raw hex to every caller")
			}
			if *op != c.want {
				t.Fatalf("got %v, want %v", *op, c.want)
			}
		})
	}
}

// Two field widths in this schema are worth asserting rather than eyeballing.
func TestHipoFinanceFieldWidths(t *testing.T) {
	decode := func(t *testing.T, body string, ext bool) any {
		t.Helper()
		raw, err := hex.DecodeString(body)
		if err != nil {
			t.Fatal(err)
		}
		cells, err := boc.DeserializeBoc(raw)
		if err != nil {
			t.Fatal(err)
		}
		var v any
		if ext {
			_, _, v, err = ExtInMessageDecoder(cells[0], nil)
		} else {
			_, _, v, err = InternalMessageDecoder(cells[0], nil)
		}
		if err != nil {
			t.Fatal(err)
		}
		return v
	}

	// borrower_reward_share is 16 bits out of 65535, not 8 out of 255. Declared as uint8 it
	// still parses - it just yields the high byte, so this borrower's 1799 reads as 7.
	loan, ok := decode(t, "B5EE9C720101030100AF00013E36335DA9000000006AA5512B6AA64F087016BFFF2242C0052D7FFFD14007070101901781DB04A92CB9444B3043E9C8DB1BB6FC23B7DA38B9E7902A8ECCDEA2632F796AA64F08000300000F130A89779CC0D6E9CBA7571982050AF2AD2C59206879E5762DC21235D3BA66020080BFF1F56BB51083DB7B22CF827282C6D759FEBE75B5D011DEC36EF29933D7AC763757B03237DA9F88F85982223F4F91C9801C33631536342937184DD93D46FE06", false).(HipoFinanceRequestLoanMsgBody)
	if !ok {
		t.Fatalf("request_loan did not decode to its body type")
	}
	if uint64(loan.BorrowerRewardShare) != 1799 {
		t.Errorf("borrower_reward_share = %d, want 1799", loan.BorrowerRewardShare)
	}

	// finish_participation is the one op here whose query_id is 32 bits. Declared as uint64
	// it swallows round_since, and the round the message is about is lost.
	finish, ok := decode(t, "B5EE9C7201010101001200002023274435000000006A95CF676A944F08", true).(HipoFinanceFinishParticipationExtInMsgBody)
	if !ok {
		t.Fatalf("finish_participation did not decode to its body type")
	}
	if finish.RoundSince != 1788202855 {
		t.Errorf("round_since = %d, want 1788202855", finish.RoundSince)
	}
}
