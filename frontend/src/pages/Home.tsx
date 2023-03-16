import { HomeBackground, NavbarBackground } from '../components/home/background'
import { SignBackground } from '../components/sign_in/background'
import logo from '../assets/logo/Carbs_light.svg'
import wallet from '../assets/logo/Wallet_duotone_line.svg'
import download from '../assets/logo/Load_circle.svg'
import lable from '../assets/logo/lable_duotone.svg'
import paper from '../assets/logo/Paper_duotone_line.svg'
import user from '../assets/logo/User_circle.svg'
import stats from '../assets/logo/stock-market-blue.png'
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
import { SmallBlackButton } from '../components/atoms/button'
import { useState } from 'react'

export const Home = () => {
    let popup
    const [showPopup, setShowPopup] = useState(false)

    const handleClick = () => {
        setShowPopup(false)
    }

    if (showPopup) {
        popup = <></>
    }

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
                                        <SmallBlackButton>Sell</SmallBlackButton>
                                    </AddSellContainer>
                                </StocksComponentContainer>
                                <StocksComponentContainer>
                                    <Header4>$$</Header4>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton>Buy</SmallBlackButton>
                                        <SmallBlackButton>Sell</SmallBlackButton>
                                    </AddSellContainer>
                                </StocksComponentContainer>
                                <StocksComponentContainer>
                                    <Header4>$$</Header4>
                                    <Header4>ABC Stocks</Header4>
                                    <AddSellContainer>
                                        <SmallBlackButton>Buy</SmallBlackButton>
                                        <SmallBlackButton>Sell</SmallBlackButton>
                                    </AddSellContainer>
                                </StocksComponentContainer>
                            </StocksContainer>
                        </StocksCardContainer>
                    </StocksCard>
                </BottomContainer>
            </HomeBackground>
        </div>
    )
}
