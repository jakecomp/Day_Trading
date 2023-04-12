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
import { AddFundsPopUp, BuyPopUp, HistoryPopUp, SellPopUp, SuccessPopUp } from '../components/popups/homePopup'
import { FieldForm, InputLabel, SignField } from '../components/sign_in/field'
import { InputComponentContainer, InputContainer } from '../components/sign_in/containers'
import {
    CloseContainer,
    HistoryComponentContainer,
    HistoryContainer,
    HistoryPopupContainer,
} from '../components/popups/containers'
import { useForm } from 'react-hook-form'
import { SignInPopUp } from '../components/popups/signinpopup'
import { SimpleLink } from '../components/atoms/links'
import { parse } from 'node:path/win32'

export const Home = (props: any) => {
    const [buyPopup, setBuyPopup] = useState(false)
    const [sellPopup, setSellPopup] = useState(false)
    const [historyPopup, setHistoryPopup] = useState(false)
    const [stockId, setStockId] = useState(0)
    const [stockList, setStockList] = useState('')
    const [addFundsPopup, setAddFundsPopup] = useState(false)
    const [SignUpPopUp, setSignUpPopup] = useState(false)
    const [SignInPopUp, setSignInPopup] = useState(true)
    const [command, setCommand] = useState('')
    const [args, setArgs] = useState('')
    const [successPopUp, setSuccessPopUp] = useState(false)

    interface commandForm {
        ticket: number
        command: string
        args1: string
        args2: FunctionStringCallback
        balance: string
    }

    const { register, handleSubmit } = useForm<commandForm>({ mode: 'onSubmit' })

    const handleSuccess = () => {
        setSuccessPopUp(false)
        setBuyPopup(false)
        setSellPopup(false)
    }
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

    const [balanceValue, setBalanceValue] = useState(0)

    const handleAdd = (data: commandForm) => {
        setSuccessPopUp(true)
        const a = balanceValue
        const b = parseInt(data.balance, 0)
        setBalanceValue(a + b)
        setAddFundsPopup(false)

        const report = {
            ticket: 0,
            command: 'ADD_FUNDS',
            args: balanceValue,
        }
        console.log(report)
        try {
            fetch('http://10.9.0.4:8000/', {
                method: 'POST',
                headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
                body: JSON.stringify(report),
            })
                .then((response) => response.text())
                .then((response) => {
                    socket = new WebSocket('ws://10.9.0.4:8000/ws?token=' + response)
                    socket.onopen = function () {
                        socket.send('Hi Hi Server')
                        socket.onmessage = (msg: any) => {
                            console.log('Server Message: ' + msg.data)
                        }
                    }
                })
        } catch (error) {
            console.error(error)
        }
    }

    let socket: WebSocket

    const RetrieveBuyData = (data: commandForm) => {
        console.log(data)
        setSuccessPopUp(true)
        setBalanceValue(balanceValue - parseInt(data.args1, 0))
        const report = {
            ticket: 0,
            command: 'COMMIT_BUY',
            args: data.args1,
        }

        console.log(report)

        try {
            fetch('http://10.9.0.4:8000/', {
                method: 'POST',
                headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
                body: JSON.stringify(report),
            })
                .then((response) => response.text())
                .then((response) => {
                    socket = new WebSocket('ws://10.9.0.4:8000/ws?token=' + response)
                    socket.onopen = function () {
                        socket.send('Hi Hi Server')
                        socket.onmessage = (msg: any) => {
                            console.log('Server Message: ' + msg.data)
                        }
                    }
                })
        } catch (error) {
            console.error(error)
        }
    }

    const RetrieveAutoBuyData = (data: commandForm) => {
        console.log(data)
        setSuccessPopUp(true)
        const report = {
            ticket: 0,
            command: 'SET_AUTO_BUY',
            args: data.args2,
        }
        console.log(report)

        try {
            fetch('http://10.9.0.4:8000/', {
                method: 'POST',
                headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
                body: JSON.stringify(report),
            })
                .then((response) => response.text())
                .then((response) => {
                    socket = new WebSocket('ws://10.9.0.4:8000/ws?token=' + response)
                    socket.onopen = function () {
                        socket.send('Hi Hi Server')
                        socket.onmessage = (msg: any) => {
                            console.log('Server Message: ' + msg.data)
                        }
                    }
                })
        } catch (error) {
            console.error(error)
        }
    }

    const RetrieveSellData = (data: commandForm) => {
        console.log(data)
        setSuccessPopUp(true)
        setBalanceValue(balanceValue + parseInt(data.args1, 0))
        const report = {
            ticket: 0,
            command: 'COMMIT_SELL',
            args: data.args1,
        }
        console.log(report)

        try {
            fetch('http://10.9.0.4:8000/', {
                method: 'POST',
                headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
                body: JSON.stringify(report),
            })
                .then((response) => response.text())
                .then((response) => {
                    socket = new WebSocket('ws://10.9.0.4:8000/ws?token=' + response)
                    socket.onopen = function () {
                        socket.send('Hi Hi Server')
                        socket.onmessage = (msg: any) => {
                            console.log('Server Message: ' + msg.data)
                        }
                    }
                })
        } catch (error) {
            console.error(error)
        }
    }

    const RetrieveAutoSellData = (data: commandForm) => {
        console.log(data)
        setSuccessPopUp(true)
        const report = {
            ticket: 0,
            command: 'SET_AUTO_SELL',
            args: data.args2,
        }
        console.log(report)

        try {
            fetch('http://10.9.0.4:8000/home', {
                method: 'POST',
                headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
                body: JSON.stringify(report),
            })
                .then((response) => response.text())
                .then((response) => {
                    socket = new WebSocket('ws://10.9.0.4:8000/ws?token=' + response)
                    socket.onopen = function () {
                        socket.send('Hi Hi Server')
                        socket.onmessage = (msg: any) => {
                            console.log('Server Message: ' + msg.data)
                        }
                    }
                })
        } catch (error) {
            console.error(error)
        }
    }

    // User Data
    const retrievedName = 'JSmith'
    const tradingBalanceValue = 0
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

    const userHistoryList = [
        'Adidas',
        'Roots',
        'RBC',
        'Nike',
        'Sub',
        'Pub',
        'Apple',
        'Emia',
        'Banshee',
        'Bayern',
        'Mercedes',
        'Samsung',
        'Freeport',
        'Nord',
    ]
    const historyLength = userHistoryList.length
    const historyLengthFlag = historyLength >= 10 ? '210px' : 'unset'
    const listHistory = userHistoryList.map((value: any, index: number) => (
        <HistoryComponentContainer key={value}>
            <Header4>
                {index + 1}.&nbsp;{userHistoryList[index]}
            </Header4>
        </HistoryComponentContainer>
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
                            <DataValue>$ {balanceValue}</DataValue>
                            <DataName>Account Balance</DataName>
                        </DataTextContainer>
                    </DataContainer>
                    <MediumBlackButton onClick={() => setAddFundsPopup(true)}>Add Funds</MediumBlackButton>
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
                        <InputLabel>Buy Stocks</InputLabel>
                        <InputPopupContainer>
                            <SignField placeholder='Set amount' {...register('args1')}></SignField>
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
                            <SignField placeholder='Set amount' {...register('args2')}></SignField>
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
                            <SignField placeholder='Set amount' {...register('args1')}></SignField>
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
                            <SignField placeholder='Set amount' {...register('args2')}></SignField>
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
                <HistoryPopupContainer>
                    <Header3>History</Header3>
                    <HistoryContainer style={{ height: historyLengthFlag }}>{listHistory}</HistoryContainer>
                </HistoryPopupContainer>
            </HistoryPopUp>

            <SuccessPopUp trigger={successPopUp}>
                <Header3>Transaction complete!</Header3>
                <BigBlackButton style={{ width: '300px' }} onClick={handleSuccess}>
                    OK
                </BigBlackButton>
            </SuccessPopUp>

            <AddFundsPopUp trigger={addFundsPopup}>
                <CloseContainer>
                    <img src={exit} style={{ width: '40px' }} onClick={() => setAddFundsPopup(false)} />
                </CloseContainer>
                <Header3> ADD FUNDS</Header3>
                <InputContainer>
                    <InputComponentContainer onSubmit={handleSubmit(handleAdd)}>
                        <InputLabel>Add funds to your account</InputLabel>
                        <InputPopupContainer>
                            <SignField placeholder='Set amount' {...register('balance')}></SignField>
                            <BigBlackButton style={{ width: '100px', height: '40px' }}>Add</BigBlackButton>
                        </InputPopupContainer>
                    </InputComponentContainer>
                </InputContainer>
            </AddFundsPopUp>
        </div>
    )
}
