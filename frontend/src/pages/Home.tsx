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
} from '../components/home/containers'
import { StatusCard, StocksCard, TradesCard } from '../components/home/card'
import { Header1, Header2, Header3, Header4 } from '../components/atoms/fonts'
import { BigBlackButton, SmallBlackButton } from '../components/atoms/button'
import { useState } from 'react'
import { BuyPopUp, SellPopUp } from '../components/popups/homePopup'
import { InputLabel, SignField } from '../components/sign_in/field'
import { InputContainer } from '../components/sign_in/containers'
import { CloseContainer } from '../components/popups/containers'

export const Home = () => {
    const [buyPopup, setBuyPopup] = useState(false)
    const [sellPopup, setSellPopup] = useState(false)

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
                            <Header2>John Smither</Header2>
                            <img src={user} />
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
                            <TradesContainer>
                                <TradesComponentContainer>
                                    <Header4>$$</Header4>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton onClick={() => setBuyPopup(true)}>Buy</SmallBlackButton>
                                        <SmallBlackButton onClick={() => setSellPopup(true)}>Sell</SmallBlackButton>
                                    </AddSellContainer>
                                </TradesComponentContainer>
                                <TradesComponentContainer>
                                    <Header4>$$</Header4>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton>Buy</SmallBlackButton>
                                        <SmallBlackButton>Sell</SmallBlackButton>
                                    </AddSellContainer>
                                </TradesComponentContainer>
                                <TradesComponentContainer>
                                    <Header4>$$</Header4>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton>Buy</SmallBlackButton>
                                        <SmallBlackButton>Sell</SmallBlackButton>
                                    </AddSellContainer>
                                </TradesComponentContainer>
                                <TradesComponentContainer>
                                    <Header4>$$</Header4>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton>Buy</SmallBlackButton>
                                        <SmallBlackButton>Sell</SmallBlackButton>
                                    </AddSellContainer>
                                </TradesComponentContainer>
                            </TradesContainer>
                            <img src={stats} width='571' height='258' />
                        </TradesCardContainer>
                    </TradesCard>
                    <StocksCard>
                        <StocksCardContainer>
                            <Header3>Available Stocks</Header3>
                            <StocksContainer>
                                <StocksComponentContainer>
                                    <Header4>$$</Header4>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton>Buy</SmallBlackButton>
                                    </AddSellContainer>
                                </StocksComponentContainer>
                                <StocksComponentContainer>
                                    <Header4>$$</Header4>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton>Buy</SmallBlackButton>
                                    </AddSellContainer>
                                </StocksComponentContainer>
                                <StocksComponentContainer>
                                    <Header4>$$</Header4>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton>Buy</SmallBlackButton>
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
                    <InputLabel>Buy more</InputLabel>
                    <SignField></SignField>
                    <InputLabel>Automatic Buy</InputLabel>
                    <SignField placeholder='Set amount'></SignField>
                </InputContainer>
                <BigBlackButton style={{ width: '100px', height: '40px' }} onClick={() => setBuyPopup(false)}>
                    Buy
                </BigBlackButton>
            </BuyPopUp>
            <SellPopUp trigger={sellPopup}>
                <CloseContainer>
                    <img src={exit} style={{ width: '40px' }} onClick={() => setSellPopup(false)} />
                </CloseContainer>
                <Header3>Stock abc sell</Header3>
                <InputContainer>
                    <InputLabel>Sell stocks</InputLabel>
                    <SignField></SignField>
                    <InputLabel>Automatic Sell</InputLabel>
                    <SignField placeholder='Set amount'></SignField>
                </InputContainer>
                <BigBlackButton style={{ width: '100px', height: '40px' }} onClick={() => setSellPopup(false)}>
                    Buy
                </BigBlackButton>
            </SellPopUp>
        </div>
    )
}
