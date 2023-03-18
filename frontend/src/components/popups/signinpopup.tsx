import { BigBlackButton } from '../atoms/button'
import { SimpleLink } from '../atoms/links'
import { SignInHeader } from '../sign_in/text'
import { PopupBackground } from './background'
import { LoggedCard } from './card'

export const SignInPopUp = (props: any) => {
    return props.trigger ? (
        <div>
            <PopupBackground>
                <LoggedCard>{props.children}</LoggedCard>
            </PopupBackground>
        </div>
    ) : null
}
