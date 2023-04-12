import { BigBlackButton } from '../atoms/button'
import { Header3 } from '../atoms/fonts'
import { SimpleLink } from '../atoms/links'
import { PopupBackground } from './background'
import { LoggedCard } from './card'
import { SignInPopUp } from './signinpopup'

export const BuyPopUp = (props: any) => {
    return props.trigger ? (
        <div>
            <PopupBackground>
                <LoggedCard>{props.children}</LoggedCard>
            </PopupBackground>
        </div>
    ) : null
}

export const SellPopUp = (props: any) => {
    return props.trigger ? (
        <div>
            <PopupBackground>
                <LoggedCard>{props.children}</LoggedCard>
            </PopupBackground>
        </div>
    ) : null
}

export const HistoryPopUp = (props: any) => {
    return props.trigger ? (
        <div>
            <PopupBackground>
                <LoggedCard>{props.children}</LoggedCard>
            </PopupBackground>
        </div>
    ) : null
}

export const SuccessPopUp = (props: any) => {
    return props.trigger ? (
        <div>
            <PopupBackground>
                <LoggedCard>{props.children}</LoggedCard>
            </PopupBackground>
        </div>
    ) : null
}

export const AddFundsPopUp = (props: any) => {
    return props.trigger ? (
        <div>
            <PopupBackground>
                <LoggedCard>{props.children}</LoggedCard>
            </PopupBackground>
        </div>
    ) : null
}
