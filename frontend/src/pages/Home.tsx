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

export const Home = (props: any) => {
    const [buyPopup, setBuyPopup] = useState(false)
    const [sellPopup, setSellPopup] = useState(false)
    const [historyPopup, setHistoryPopup] = useState(false)
    const [stockId, setStockId] = useState(0)
    const [stockList, setStockList] = useState('')

    interface valueForm {
        buy: number
        autobuy: number
        sell: number
        autoSell: number
    }

    const { register, handleSubmit } = useForm<valueForm>({ mode: 'onSubmit' })

    const handleBuy = (index: number, list: any) => {
        setBuyPopup(true)
        setStockId(index)
        setStockList(list)
    }

    const handleSell = (index: number, list: any) => {
        setSellPopup(true)
        setStockId(index)
        setStockList(list)
    }

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

    // User Data
    const retrievedName = 'Anita B. Etin'
    const tradingBalanceValue = '$5,291'
    const investmentValue = '$7,042'
    const rateofReturnValue = '148%'
    const numberOfTradesValue = '87'

    const availableStocks = [
        'Lululemon',
        'Netflix',
        'Air Canada',
        'Amazon',
        'Luke',
        'Emia',
        'Banshee',
        'Bayern',
        'Mercedes',
        'Samsung',
        'Freeport',
        'Nord',
    ]
    const listStocks = availableStocks.length >= 10 ? '388px' : 'unset'

    const stocksList = availableStocks.map((value: any, index: any) => (
        <StocksComponentContainer key={value}>
            <Header4>
                {index + 1}.&nbsp;{availableStocks[index]}
            </Header4>
            <AddSellContainer>
                <SmallBlackButton onClick={() => handleBuy(index, availableStocks)}>Buy</SmallBlackButton>
            </AddSellContainer>
        </StocksComponentContainer>
    ))

    const userTradesList = ['Adidas', 'Roots', 'RBC', 'Nike', 'Sub', 'Pub', 'Apple']
    const tradesLength = userTradesList.length
    const tradesLengthFlag = tradesLength >= 5 ? '190px' : 'unset'

    const listTrades = userTradesList.map((value: any, index: number) => (
        <TradesComponentContainer key={value}>
            <Header4>
                {index + 1}.&nbsp;{userTradesList[index]}
            </Header4>
            <AddSellContainer>
                <SmallBlackButton onClick={() => handleBuy(index, userTradesList)}>Buy</SmallBlackButton>
                <SmallBlackButton onClick={() => handleSell(index, userTradesList)}>Sell</SmallBlackButton>
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
                                <Header2>{retrievedName}</Header2>
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
                            <DataValue>{tradingBalanceValue}</DataValue>
                            <DataName>Trading Balance</DataName>
                        </DataTextContainer>
                    </DataContainer>
                    <DataContainer>
                        <img src={download} />
                        <DataTextContainer>
                            <DataValue>{investmentValue}</DataValue>
                            <DataName>Investment</DataName>
                        </DataTextContainer>
                    </DataContainer>
                    <DataContainer>
                        <img src={lable} />
                        <DataTextContainer>
                            <DataValue>{rateofReturnValue}</DataValue>
                            <DataName>Rate of Return</DataName>
                        </DataTextContainer>
                    </DataContainer>
                    <DataContainer>
                        <img src={paper} />
                        <DataTextContainer>
                            <DataValue>{numberOfTradesValue}</DataValue>
                            <DataName>Number of Trades</DataName>
                        </DataTextContainer>
                    </DataContainer>
                </StatusCard>
                <BottomContainer>
                    <TradesCard>
                        <TradesCardContainer>
                            <Header3>My Trades</Header3>
                            <TradesContainer style={{ height: tradesLengthFlag }}>{listTrades}</TradesContainer>
                            <img src={stats} width='400' height='200' />
                        </TradesCardContainer>
                    </TradesCard>
                    <StocksCard>
                        <StocksCardContainer>
                            <Header3>Available Stocks</Header3>
                            <StocksContainer style={{ height: listStocks }}>{stocksList}</StocksContainer>
                        </StocksCardContainer>
                    </StocksCard>
                </BottomContainer>
            </HomeBackground>

            <BuyPopUp trigger={buyPopup}>
                <CloseContainer>
                    <img src={exit} style={{ width: '40px' }} onClick={() => setBuyPopup(false)} />
                </CloseContainer>
                <Header3>{stockList[stockId]} BUY</Header3>
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
                <Header3>{stockList[stockId]} SELL</Header3>
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
