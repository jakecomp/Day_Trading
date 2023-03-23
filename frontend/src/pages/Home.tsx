import { HomeBackground, NavbarBackground } from '../components/home/background'
import { SignBackground } from '../components/sign_in/background'
import logo from '../assets/logo/Carbs_light.svg'
import wallet from '../assets/logo/Wallet_duotone_line.svg'
import download from '../assets/logo/Load_circle.svg'
import lable from '../assets/logo/lable_duotone.svg'
import paper from '../assets/logo/Paper_duotone_line.svg'
import user from '../assets/logo/User_circle.svg'
import stats from '../assets/logo/stock-market-blue.png'
import exit from '../assets/logo/Close_round_light.svg'
import {
    DataContainer,
    DataTextContainer,
    DataValue,
    UserContainer,
    UserTextContainer,
    DataName,
    StickyContainer,
    BottomContainer,
    TradesComponentContainer,
    AddSellContainer,
    TradesContainer,
    TradesCardContainer,
    StocksCardContainer,
    StocksContainer,
    StocksComponentContainer,
    InputPopupContainer,
    AccountContainer,
} from '../components/home/containers'
import { StatusCard, StocksCard, TradesCard } from '../components/home/card'
import { Header1, Header2, Header3, Header4 } from '../components/atoms/fonts'
import { BigBlackButton, MediumBlackButton, SmallBlackButton } from '../components/atoms/button'
import { useState } from 'react'
import { BuyPopUp, HistoryPopUp, SellPopUp } from '../components/popups/homePopup'
import { FieldForm, InputLabel, SignField } from '../components/sign_in/field'
import { InputComponentContainer, InputContainer } from '../components/sign_in/containers'
import { CloseContainer } from '../components/popups/containers'
import { useForm } from 'react-hook-form'

export default function GetStocks(props: any) {
    const data = props.data
    const length = data.length

    // const showList = data.map((item: any, index: any) => (
    //     <TradesComponentContainer key={index} value={item}>
    //         {item}
    //     </TradesComponentContainer>
    // ))
}

export const Home = (props: any) => {
    const [buyPopup, setBuyPopup] = useState(false)
    const [sellPopup, setSellPopup] = useState(false)
    const [historyPopup, setHistoryPopup] = useState(false)

    interface valueForm {
        buy: number
        autobuy: number
        sell: number
        autoSell: number
    }

    const { register, handleSubmit } = useForm<valueForm>({ mode: 'onSubmit' })

    const RetrieveBuyData = (data: valueForm) => {
        console.log('Buy')
        console.log(data)
    }

    const RetrieveAutoBuyData = (data: valueForm) => {
        console.log('AutoBuy')
        console.log(data)
    }

    const RetrieveSellData = (data: valueForm) => {
        console.log('Sell')
        console.log(data)
    }

    const RetrieveAutoSellData = (data: valueForm) => {
        console.log('AutoSell')
        console.log(data)
    }

    const data = ['Adidas', 'Roots', 'RBC', 'Nike']

    const listNumbers = data.map((numbers: any, index: any) => (
        <TradesComponentContainer key={numbers}>
            <Header4>
                {index + 1}.&nbsp;{data[index]}
            </Header4>
            <AddSellContainer>
                <SmallBlackButton onClick={() => setBuyPopup(true)}>Buy</SmallBlackButton>
                <SmallBlackButton onClick={() => setSellPopup(true)}>Sell</SmallBlackButton>
            </AddSellContainer>
        </TradesComponentContainer>
    ))

    return (
        <div>
            <StickyContainer>
                <NavbarBackground>
                    <UserTextContainer>
                        <img src={logo} />
                        <Header1>DTA</Header1>
                    </UserTextContainer>
                    <UserContainer>
                        <UserTextContainer>
                            <MediumBlackButton onClick={() => setHistoryPopup(true)}>History</MediumBlackButton>
                            <AccountContainer>
                                <Header2>John Smither</Header2>
                                <img src={user} />
                            </AccountContainer>
                        </UserTextContainer>
                    </UserContainer>
                </NavbarBackground>
            </StickyContainer>

            <HomeBackground>
                <StatusCard>
                    <DataContainer>
                        <img src={wallet} />
                        <DataTextContainer>
                            <DataValue>$5,912</DataValue>
                            <DataName>Trading Balance</DataName>
                        </DataTextContainer>
                    </DataContainer>
                    <DataContainer>
                        <img src={download} />
                        <DataTextContainer>
                            <DataValue>$8,426</DataValue>
                            <DataName>Investment</DataName>
                        </DataTextContainer>
                    </DataContainer>
                    <DataContainer>
                        <img src={lable} />
                        <DataTextContainer>
                            <DataValue>185%</DataValue>
                            <DataName>Rate of Return</DataName>
                        </DataTextContainer>
                    </DataContainer>
                    <DataContainer>
                        <img src={paper} />
                        <DataTextContainer>
                            <DataValue>419</DataValue>
                            <DataName>Number of Trades</DataName>
                        </DataTextContainer>
                    </DataContainer>
                </StatusCard>
                <BottomContainer>
                    <TradesCard>
                        <TradesCardContainer>
                            <Header3>My Trades</Header3>
                            <TradesContainer>{listNumbers}</TradesContainer>
                            <img src={stats} width='571' height='258' />
                        </TradesCardContainer>
                    </TradesCard>
                    <StocksCard>
                        <StocksCardContainer>
                            <Header3>Available Stocks</Header3>
                            <StocksContainer>
                                <StocksComponentContainer>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton onClick={() => setBuyPopup(true)}>Buy</SmallBlackButton>
                                    </AddSellContainer>
                                </StocksComponentContainer>
                            </StocksContainer>
                        </StocksCardContainer>
                    </StocksCard>
                </BottomContainer>
            </HomeBackground>

            <BuyPopUp trigger={buyPopup}>
                <CloseContainer>
                    <img src={exit} style={{ width: '40px' }} onClick={() => setBuyPopup(false)} />
                </CloseContainer>
                <Header3>Stock abc buy</Header3>
                <InputContainer>
                    <InputComponentContainer onSubmit={handleSubmit(RetrieveBuyData)}>
                        <InputLabel>Buy more</InputLabel>
                        <InputPopupContainer>
                            <SignField {...register('buy')}></SignField>
                            <BigBlackButton
                                style={{ width: '100px', height: '40px' }}
                                // onClick={() => setBuyPopup(false)}
                            >
                                Buy
                            </BigBlackButton>
                        </InputPopupContainer>
                    </InputComponentContainer>

                    <InputComponentContainer onSubmit={handleSubmit(RetrieveAutoBuyData)}>
                        <InputLabel>Automatic Buy</InputLabel>
                        <InputPopupContainer>
                            <SignField placeholder='Set amount' {...register('autobuy')}></SignField>
                            <BigBlackButton
                                style={{ width: '100px', height: '40px' }}
                                // onClick={() => setBuyPopup(false)}
                            >
                                Buy
                            </BigBlackButton>
                        </InputPopupContainer>
                    </InputComponentContainer>
                </InputContainer>
            </BuyPopUp>

            <SellPopUp trigger={sellPopup}>
                <CloseContainer>
                    <img src={exit} style={{ width: '40px' }} onClick={() => setSellPopup(false)} />
                </CloseContainer>
                <Header3>Stock abc sell</Header3>
                <InputContainer>
                    <InputComponentContainer onSubmit={handleSubmit(RetrieveSellData)}>
                        <InputLabel>Sell stocks</InputLabel>
                        <InputPopupContainer>
                            <SignField {...register('sell')}></SignField>
                            <BigBlackButton
                                style={{ width: '100px', height: '40px' }}
                                // onClick={() => setSellPopup(false)}
                            >
                                Sell
                            </BigBlackButton>
                        </InputPopupContainer>
                    </InputComponentContainer>

                    <InputComponentContainer onSubmit={handleSubmit(RetrieveAutoSellData)}>
                        <InputLabel>Automatic Sell</InputLabel>
                        <InputPopupContainer>
                            <SignField placeholder='Set amount' {...register('autoSell')}></SignField>
                            <BigBlackButton
                                style={{ width: '100px', height: '40px' }}
                                // onClick={() => setSellPopup(false)}
                            >
                                Sell
                            </BigBlackButton>
                        </InputPopupContainer>
                    </InputComponentContainer>
                </InputContainer>
            </SellPopUp>

            <HistoryPopUp trigger={historyPopup}>
                <CloseContainer>
                    <img src={exit} style={{ width: '40px' }} onClick={() => setHistoryPopup(false)} />
                </CloseContainer>
            </HistoryPopUp>
        </div>
    )
}
